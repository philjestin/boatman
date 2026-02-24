// Package eventbus provides a NATS-backed event bus with persistence to EventStore.
package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"

	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

// Bus wraps an embedded NATS server, a NATS client, and an EventStore for persistence.
type Bus struct {
	server     *server.Server
	conn       *nats.Conn
	eventStore storage.EventStore

	mu   sync.Mutex
	subs []*nats.Subscription
}

// Option configures the Bus.
type Option func(*busConfig)

type busConfig struct {
	port       int
	dataDir    string
	eventStore storage.EventStore
}

// WithPort sets the NATS server port. 0 or -1 means random available port.
func WithPort(port int) Option {
	return func(c *busConfig) { c.port = port }
}

// WithDataDir sets the NATS data directory for persistence.
func WithDataDir(dir string) Option {
	return func(c *busConfig) { c.dataDir = dir }
}

// WithEventStore attaches an EventStore for event persistence.
func WithEventStore(es storage.EventStore) Option {
	return func(c *busConfig) { c.eventStore = es }
}

// New creates and starts a new Bus with an embedded NATS server.
func New(opts ...Option) (*Bus, error) {
	cfg := &busConfig{
		port: -1, // random port by default
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Configure embedded NATS server
	natsOpts := &server.Options{
		Port:     cfg.port,
		NoLog:    true,
		NoSigs:   true,
		MaxPayload: 8 * 1024 * 1024, // 8MB
	}
	if cfg.dataDir != "" {
		natsOpts.StoreDir = cfg.dataDir
	}

	ns, err := server.NewServer(natsOpts)
	if err != nil {
		return nil, fmt.Errorf("create nats server: %w", err)
	}

	// Start server in background
	go ns.Start()

	// Wait for server to be ready
	if !ns.ReadyForConnections(5 * time.Second) {
		ns.Shutdown()
		return nil, fmt.Errorf("nats server failed to start")
	}

	// Connect client
	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		ns.Shutdown()
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	return &Bus{
		server:     ns,
		conn:       nc,
		eventStore: cfg.eventStore,
	}, nil
}

// Publish sends an event to the bus and persists it if an EventStore is configured.
func (b *Bus) Publish(ctx context.Context, event *storage.Event) error {
	// Persist first
	if b.eventStore != nil {
		if err := b.eventStore.Publish(ctx, event); err != nil {
			return fmt.Errorf("persist event: %w", err)
		}
	}

	// Build subject
	subject := BuildSubject(event.Scope.OrgID, event.Scope.TeamID, event.Type)

	// Marshal and publish
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if err := b.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("publish to nats: %w", err)
	}

	return b.conn.Flush()
}

// Subscribe returns a channel of events matching the given subject pattern.
// The returned cancel function stops the subscription and closes the channel.
func (b *Bus) Subscribe(_ context.Context, subject string) (<-chan *storage.Event, func(), error) {
	ch := make(chan *storage.Event, 64)

	sub, err := b.conn.Subscribe(subject, func(msg *nats.Msg) {
		var event storage.Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return
		}
		select {
		case ch <- &event:
		default:
			// Drop if channel is full
		}
	})
	if err != nil {
		close(ch)
		return nil, nil, fmt.Errorf("subscribe: %w", err)
	}

	b.mu.Lock()
	b.subs = append(b.subs, sub)
	b.mu.Unlock()

	cancel := func() {
		sub.Unsubscribe()
		close(ch)
	}

	return ch, cancel, nil
}

// Replay queries persisted events matching the filter and returns them via a channel.
func (b *Bus) Replay(ctx context.Context, filter storage.EventFilter) (<-chan *storage.Event, error) {
	if b.eventStore == nil {
		return nil, fmt.Errorf("replay requires an EventStore")
	}

	events, err := b.eventStore.Query(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("query events for replay: %w", err)
	}

	ch := make(chan *storage.Event, len(events))
	go func() {
		defer close(ch)
		for _, e := range events {
			select {
			case ch <- e:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// Close shuts down the bus, draining connections and stopping the server.
func (b *Bus) Close() error {
	b.mu.Lock()
	for _, sub := range b.subs {
		sub.Unsubscribe()
	}
	b.subs = nil
	b.mu.Unlock()

	if b.conn != nil {
		b.conn.Close()
	}
	if b.server != nil {
		b.server.Shutdown()
	}
	return nil
}

// ClientURL returns the NATS client connection URL (useful for testing).
func (b *Bus) ClientURL() string {
	return b.server.ClientURL()
}

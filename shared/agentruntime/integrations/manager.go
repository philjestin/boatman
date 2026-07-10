package integrations

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

// Connection is a reusable integration handle managed outside a single model
// run. The first implementation is descriptor-backed; future versions can hold
// live MCP clients, OAuth refreshers, or service-specific health probes behind
// the same shape.
type Connection struct {
	Name        string                     `json:"name"`
	HandleID    string                     `json:"handleId,omitempty"`
	State       State                      `json:"state"`
	Status      Status                     `json:"status"`
	MCP         *agentruntime.MCPServerRef `json:"mcp,omitempty"`
	ConnectedAt *time.Time                 `json:"connectedAt,omitempty"`
	UpdatedAt   time.Time                  `json:"updatedAt"`
}

// Manager owns long-lived integration connection state.
type Manager struct {
	Broker *Broker
	Now    func() time.Time

	mu          sync.Mutex
	connections map[string]Connection
}

// NewManager creates an integration connection manager.
func NewManager(catalog Catalog) *Manager {
	return &Manager{
		Broker:      NewBroker(catalog),
		connections: make(map[string]Connection),
	}
}

// Connect resolves configuration and stores a reusable connection handle when
// the integration is ready.
func (m *Manager) Connect(ctx context.Context, name string, opts ResolveOptions) (Connection, error) {
	if err := ctx.Err(); err != nil {
		return Connection{}, err
	}
	if m == nil {
		return Connection{}, fmt.Errorf("integration manager is nil")
	}
	broker := m.broker()
	ref, status, err := broker.ResolveMCP(ctx, name, opts)
	now := m.now()
	if err != nil && status.Name == "" {
		return Connection{}, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	existing := m.connections[status.Name]
	conn := Connection{
		Name:      status.Name,
		State:     status.State,
		Status:    status,
		UpdatedAt: now,
	}
	if status.State == StateReady || status.State == StateConnected {
		status.State = StateConnected
		status.Message = "integration connection handle is ready"
		status.Metadata = mergeStatusMetadata(status.Metadata, map[string]string{
			"transport": transportForRef(ref),
		})
		handleID := existing.HandleID
		connectedAt := existing.ConnectedAt
		if handleID == "" {
			handleID = fmt.Sprintf("%s-%d", status.Name, now.UnixNano())
			t := now
			connectedAt = &t
		}
		status.Metadata["handle_id"] = handleID
		conn = Connection{
			Name:        status.Name,
			HandleID:    handleID,
			State:       StateConnected,
			Status:      status,
			MCP:         &ref,
			ConnectedAt: connectedAt,
			UpdatedAt:   now,
		}
	} else if existing.HandleID != "" {
		conn.HandleID = existing.HandleID
		conn.ConnectedAt = existing.ConnectedAt
	}
	if err != nil {
		conn.Status.State = StateFailed
		conn.Status.Message = err.Error()
		conn.State = StateFailed
	}
	m.connections[conn.Name] = conn
	return conn, err
}

// Disconnect removes a cached connection handle.
func (m *Manager) Disconnect(ctx context.Context, name string) (Connection, error) {
	if err := ctx.Err(); err != nil {
		return Connection{}, err
	}
	if m == nil {
		return Connection{}, fmt.Errorf("integration manager is nil")
	}
	name = normalizeName(name)
	if name == "" {
		return Connection{}, fmt.Errorf("integration name is required")
	}
	now := m.now()
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.connections[name]
	if !ok {
		status := Status{Name: name, State: StateDisabled, Message: "integration connection is not active", LastChecked: now}
		conn := Connection{Name: name, State: StateDisabled, Status: status, UpdatedAt: now}
		m.connections[name] = conn
		return conn, nil
	}
	existing.State = StateDisabled
	existing.Status.State = StateDisabled
	existing.Status.Message = "integration connection disconnected"
	existing.Status.LastChecked = now
	existing.Status.Metadata = mergeStatusMetadata(existing.Status.Metadata, map[string]string{"disconnected": "true"})
	existing.MCP = nil
	existing.UpdatedAt = now
	m.connections[name] = existing
	return existing, nil
}

// List returns cached connections in deterministic name order.
func (m *Manager) List() []Connection {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	conns := make([]Connection, 0, len(m.connections))
	for _, conn := range m.connections {
		conns = append(conns, conn)
	}
	sort.Slice(conns, func(i, j int) bool {
		return conns[i].Name < conns[j].Name
	})
	return conns
}

// Event converts a connection state into an integration.state runtime event.
func (m *Manager) Event(conn Connection) agentruntime.Event {
	return m.broker().Event(conn.Status)
}

func (m *Manager) broker() *Broker {
	if m != nil && m.Broker != nil {
		return m.Broker
	}
	return NewBroker(DefaultCatalog())
}

func (m *Manager) now() time.Time {
	if m != nil && m.Now != nil {
		return m.Now().UTC()
	}
	if m != nil && m.Broker != nil && m.Broker.Now != nil {
		return m.Broker.Now().UTC()
	}
	return time.Now().UTC()
}

func transportForRef(ref agentruntime.MCPServerRef) string {
	if ref.URL != "" {
		return "remote_mcp"
	}
	if ref.Command != "" {
		return "local_mcp"
	}
	return "descriptor"
}

func mergeStatusMetadata(base, extra map[string]string) map[string]string {
	out := cloneStringMap(base)
	if out == nil {
		out = make(map[string]string)
	}
	for key, value := range extra {
		out[key] = value
	}
	return out
}

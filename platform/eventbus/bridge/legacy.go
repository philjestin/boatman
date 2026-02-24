package bridge

import (
	"context"
	"encoding/json"
	"io"

	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
	sharedevents "github.com/philjestin/boatman-ecosystem/shared/events"
)

// LegacyBridge subscribes to the event bus and writes events as stdout JSON
// in the existing shared/events.Event format, preserving backward compatibility
// with the desktop subprocess integration.
type LegacyBridge struct {
	bus    *eventbus.Bus
	writer io.Writer
	cancel func()
}

// NewLegacyBridge creates a bridge that writes bus events to the given writer
// in the legacy shared/events.Event JSON format.
func NewLegacyBridge(bus *eventbus.Bus, writer io.Writer, subject string) (*LegacyBridge, error) {
	ch, cancel, err := bus.Subscribe(context.Background(), subject)
	if err != nil {
		return nil, err
	}

	lb := &LegacyBridge{
		bus:    bus,
		writer: writer,
		cancel: cancel,
	}

	go lb.pump(ch)

	return lb, nil
}

func (lb *LegacyBridge) pump(ch <-chan *storage.Event) {
	enc := json.NewEncoder(lb.writer)
	for event := range ch {
		// Convert to legacy format
		legacy := toLegacyEvent(event)
		enc.Encode(legacy)
	}
}

// Close stops the bridge subscription.
func (lb *LegacyBridge) Close() {
	if lb.cancel != nil {
		lb.cancel()
	}
}

// toLegacyEvent converts a platform storage.Event to the shared/events.Event format.
func toLegacyEvent(e *storage.Event) sharedevents.Event {
	return sharedevents.Event{
		Type:    e.Type,
		ID:      e.ID,
		Name:    e.Name,
		Message: e.Message,
		Data:    e.Data,
	}
}

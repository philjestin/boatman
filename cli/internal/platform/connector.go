// Package platform provides connectivity between the CLI and the Boatman platform server.
// When the platform is unreachable, the CLI falls back to standalone operation.
package platform

import (
	"context"
	"time"

	platformclient "github.com/philjestin/boatman-ecosystem/platform/client"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

// Connector manages the connection to the platform server.
type Connector struct {
	client    *platformclient.Client
	connected bool
}

// TryConnect attempts to connect to the platform server.
// It is non-blocking and best-effort: if the server is unreachable,
// connected will be false and all CLI functionality continues in standalone mode.
func TryConnect(ctx context.Context, serverURL string, scope storage.Scope) *Connector {
	if serverURL == "" {
		return &Connector{}
	}

	client := platformclient.New(serverURL, scope)

	// Quick ping with short timeout
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx); err != nil {
		return &Connector{client: client, connected: false}
	}

	return &Connector{client: client, connected: true}
}

// IsConnected returns whether the platform server is reachable.
func (c *Connector) IsConnected() bool {
	return c.connected
}

// Client returns the underlying platform client. May be nil if no server URL was configured.
func (c *Connector) Client() *platformclient.Client {
	return c.client
}

// Package contextpin provides an adapter wrapping the harness contextpin
// with coordinator integration.
package contextpin

import (
	"github.com/philjestin/boatman-ecosystem/harness/contextpin"
	"github.com/philjestin/boatmanmode/internal/coordinator"
)

// Type aliases from harness
type Pin = contextpin.Pin
type DependencyGraph = contextpin.DependencyGraph
type FileLockError = contextpin.FileLockError
type PinHandoff = contextpin.PinHandoff

// coordinatorFileLock adapts coordinator.Coordinator to the contextpin.FileLock interface.
type coordinatorFileLock struct {
	coord *coordinator.Coordinator
}

func (c *coordinatorFileLock) LockFiles(ownerID string, files []string) bool {
	return c.coord.LockFiles(ownerID, files)
}

func (c *coordinatorFileLock) UnlockFiles(ownerID string, files []string) {
	c.coord.UnlockFiles(ownerID, files)
}

// ContextPinner wraps the harness ContextPinner with coordinator integration.
type ContextPinner struct {
	*contextpin.ContextPinner
}

// New creates a new ContextPinner.
func New(worktreePath string) *ContextPinner {
	return &ContextPinner{
		ContextPinner: contextpin.New(worktreePath),
	}
}

// SetCoordinator sets the coordinator for file locking.
func (cp *ContextPinner) SetCoordinator(c *coordinator.Coordinator) {
	cp.ContextPinner.SetFileLock(&coordinatorFileLock{coord: c})
}

// NewDependencyGraph creates a new dependency graph.
var NewDependencyGraph = contextpin.NewDependencyGraph

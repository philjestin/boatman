package memory

// MemoryProvider abstracts memory retrieval and persistence.
// The existing Store satisfies this interface. Platform implementations
// can provide network-backed alternatives.
type MemoryProvider interface {
	Get(projectPath string) (*Memory, error)
	Save(mem *Memory) error
}

// Compile-time check that Store implements MemoryProvider.
var _ MemoryProvider = (*Store)(nil)

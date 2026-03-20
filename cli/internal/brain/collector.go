package brain

import (
	"path/filepath"
	"strings"
	"sync"

	harnessbrain "github.com/philjestin/boatman-ecosystem/harness/brain"
)

// Collector passively detects knowledge gap signals during agent workflows.
type Collector struct {
	store     *harnessbrain.SignalStore
	fileReads map[string]int // track repeated file reads
	mu        sync.Mutex
}

// NewCollector creates a signal collector backed by a signal store.
func NewCollector(projectPath string) (*Collector, error) {
	store, err := harnessbrain.NewSignalStore("")
	if err != nil {
		return nil, err
	}

	return &Collector{
		store:     store,
		fileReads: make(map[string]int),
	}, nil
}

// OnReviewFailure records a signal when code review finds issues.
func (c *Collector) OnReviewFailure(issues []string, filesChanged []string) {
	if len(issues) == 0 {
		return
	}

	domain := inferDomain(filesChanged)

	c.store.Record(harnessbrain.Signal{
		Type:      harnessbrain.SignalReviewFailure,
		Domain:    domain,
		Details:   strings.Join(issues, "; "),
		FilePaths: filesChanged,
	})
}

// OnRefactorIteration records a signal when repeated refactoring is needed.
func (c *Collector) OnRefactorIteration(iteration int, issues []string, filesChanged []string) {
	if iteration < 2 {
		return // First iteration is normal
	}

	domain := inferDomain(filesChanged)

	c.store.Record(harnessbrain.Signal{
		Type:      harnessbrain.SignalRefactorLoop,
		Domain:    domain,
		Details:   strings.Join(issues, "; "),
		FilePaths: filesChanged,
	})
}

// OnFileRead records file reads and emits a signal when 3+ reads of the same file occur.
func (c *Collector) OnFileRead(path string) {
	c.mu.Lock()
	c.fileReads[path]++
	count := c.fileReads[path]
	c.mu.Unlock()

	if count >= 3 {
		domain := inferDomain([]string{path})

		c.store.Record(harnessbrain.Signal{
			Type:      harnessbrain.SignalRepeatedFileRead,
			Domain:    domain,
			Details:   path + " read " + strings.Repeat(".", count) + " times",
			FilePaths: []string{path},
		})
	}
}

// Flush persists accumulated signals to disk.
func (c *Collector) Flush() error {
	return c.store.Save()
}

// inferDomain guesses a domain area from file paths.
func inferDomain(filePaths []string) string {
	for _, fp := range filePaths {
		// Check for pack-based organization: packs/<domain>/...
		if strings.HasPrefix(fp, "packs/") {
			parts := strings.SplitN(fp, "/", 3)
			if len(parts) >= 2 {
				return parts[1]
			}
		}

		// Check for engine-based organization: engines/<domain>/...
		if strings.HasPrefix(fp, "engines/") {
			parts := strings.SplitN(fp, "/", 3)
			if len(parts) >= 2 {
				return parts[1]
			}
		}

		// Fall back to directory name
		dir := filepath.Dir(fp)
		if dir != "." {
			return filepath.Base(dir)
		}
	}

	return "unknown"
}

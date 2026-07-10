package brain

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"

	harnessbrain "github.com/philjestin/boatman-ecosystem/harness/brain"
)

// Collector passively detects knowledge gap signals during agent workflows.
type Collector struct {
	store     *harnessbrain.SignalStore
	graph     *KnowledgeGraph
	taskID    string
	taskTitle string
	fileReads map[string]int // track repeated file reads
	mu        sync.Mutex
}

// NewCollector creates a signal collector backed by a signal store.
func NewCollector(projectPath string) (*Collector, error) {
	store, err := harnessbrain.NewSignalStore("")
	if err != nil {
		return nil, err
	}
	graph, _ := LoadKnowledgeGraph(projectPath)

	return &Collector{
		store:     store,
		graph:     graph,
		fileReads: make(map[string]int),
	}, nil
}

// OnTaskContext records the task and files selected during planning.
func (c *Collector) OnTaskContext(taskID, title string, filePaths []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.taskID = taskID
	c.taskTitle = title
	c.graph.RecordTaskContext(taskID, title, filePaths)
}

// OnTaskExecution records the files changed by implementation.
func (c *Collector) OnTaskExecution(taskID, title string, filesChanged []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.taskID = taskID
	c.taskTitle = title
	c.graph.RecordTaskExecution(taskID, title, filesChanged)
}

// OnReviewFailure records a signal when code review finds issues.
func (c *Collector) OnReviewFailure(issues []string, filesChanged []string) {
	if len(issues) == 0 {
		return
	}

	domain := inferDomain(filesChanged)

	c.recordSignal(harnessbrain.Signal{
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

	c.recordSignal(harnessbrain.Signal{
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

		c.recordSignal(harnessbrain.Signal{
			Type:      harnessbrain.SignalRepeatedFileRead,
			Domain:    domain,
			Details:   path + " read " + strings.Repeat(".", count) + " times",
			FilePaths: []string{path},
		})
	}
}

// OnBrainsDistilled records generated brain artifacts in the graph.
func (c *Collector) OnBrainsDistilled(results []DistillResult) {
	if len(results) == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, result := range results {
		c.graph.RecordBrain(result)
	}
}

// Flush persists accumulated signals to disk.
func (c *Collector) Flush() error {
	var errs []error
	if c.store != nil {
		errs = append(errs, c.store.Save())
	}
	if c.graph != nil {
		errs = append(errs, c.graph.Save())
	}
	return errors.Join(errs...)
}

func (c *Collector) recordSignal(sig harnessbrain.Signal) {
	if c.store != nil {
		c.store.Record(sig)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.graph != nil {
		c.graph.RecordSignal(sig)
		if c.taskID != "" || c.taskTitle != "" {
			c.graph.RecordTaskSignal(c.taskID, c.taskTitle, sig)
		}
	}
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

// Package brain provides curated domain knowledge for AI agents.
// Brains are loaded on demand when an agent encounters a matching domain,
// providing business rules, invariants, key code paths, and failure modes.
package brain

// Brain represents a curated domain knowledge document.
type Brain struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     int       `json:"version"`
	Description string    `json:"description"`
	Triggers    Triggers  `json:"triggers"`
	Confidence  float64   `json:"confidence"`
	LastUpdated string    `json:"last_updated"`
	Sections    []Section `json:"sections"`
	References  []Reference `json:"references"`
}

// Triggers defines when a brain should be loaded.
type Triggers struct {
	Keywords     []string `json:"keywords"`
	Entities     []string `json:"entities"`
	FilePatterns []string `json:"file_patterns"`
}

// Section is a named block of markdown content within a brain.
type Section struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Reference points to a file in the codebase that the brain covers.
type Reference struct {
	Path        string `json:"path"`
	Description string `json:"description"`
	Checksum    string `json:"checksum"`
}

// MatchContext provides the inputs for trigger matching.
type MatchContext struct {
	Keywords  []string
	FilePaths []string
	Entities  []string
}

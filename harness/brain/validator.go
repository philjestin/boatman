package brain

import (
	"crypto/sha256"
	"fmt"
	"os"
)

// ValidationError represents a schema or content issue with a brain.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// StaleReference indicates a brain reference whose file has changed or is missing.
type StaleReference struct {
	Reference Reference
	Reason    string // "missing", "changed"
}

// MaxSections is the maximum number of sections allowed per brain.
const MaxSections = 20

// MaxSectionContentLen is the maximum character length for a single section's content.
const MaxSectionContentLen = 10000

// Validate checks a brain for schema conformance and size budgets.
func Validate(b *Brain) []ValidationError {
	var errs []ValidationError

	if b.ID == "" {
		errs = append(errs, ValidationError{Field: "id", Message: "required"})
	}
	if b.Name == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "required"})
	}
	if b.Version < 1 {
		errs = append(errs, ValidationError{Field: "version", Message: "must be >= 1"})
	}
	if b.Confidence < 0 || b.Confidence > 1 {
		errs = append(errs, ValidationError{Field: "confidence", Message: "must be between 0.0 and 1.0"})
	}

	if len(b.Triggers.Keywords) == 0 && len(b.Triggers.Entities) == 0 && len(b.Triggers.FilePatterns) == 0 {
		errs = append(errs, ValidationError{Field: "triggers", Message: "at least one trigger type required"})
	}

	if len(b.Sections) == 0 {
		errs = append(errs, ValidationError{Field: "sections", Message: "at least one section required"})
	}
	if len(b.Sections) > MaxSections {
		errs = append(errs, ValidationError{
			Field:   "sections",
			Message: fmt.Sprintf("too many sections (%d, max %d)", len(b.Sections), MaxSections),
		})
	}

	for i, s := range b.Sections {
		if s.Title == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("sections[%d].title", i),
				Message: "required",
			})
		}
		if len(s.Content) > MaxSectionContentLen {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("sections[%d].content", i),
				Message: fmt.Sprintf("too long (%d chars, max %d)", len(s.Content), MaxSectionContentLen),
			})
		}
	}

	return errs
}

// CheckStaleness compares stored reference checksums against current files.
// projectRoot is the base directory for resolving reference paths.
func CheckStaleness(b *Brain, projectRoot string) []StaleReference {
	var stale []StaleReference

	for _, ref := range b.References {
		if ref.Checksum == "" {
			continue
		}

		fullPath := ref.Path
		if projectRoot != "" {
			fullPath = projectRoot + "/" + ref.Path
		}

		data, err := os.ReadFile(fullPath)
		if err != nil {
			stale = append(stale, StaleReference{
				Reference: ref,
				Reason:    "missing",
			})
			continue
		}

		currentChecksum := fmt.Sprintf("%x", sha256.Sum256(data))
		if currentChecksum != ref.Checksum {
			stale = append(stale, StaleReference{
				Reference: ref,
				Reason:    "changed",
			})
		}
	}

	return stale
}

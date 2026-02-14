package diff

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// DiffComment represents a comment on a diff line
type DiffComment struct {
	ID        string `json:"id"`
	LineNum   int    `json:"lineNum"`
	HunkID    string `json:"hunkId,omitempty"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	Author    string `json:"author,omitempty"`
}

// FileDiff represents a diff for a single file
type FileDiff struct {
	OldPath  string        `json:"oldPath"`
	NewPath  string        `json:"newPath"`
	Hunks    []Hunk        `json:"hunks"`
	IsNew    bool          `json:"isNew"`
	IsDelete bool          `json:"isDelete"`
	IsBinary bool          `json:"isBinary"`
	Approved bool          `json:"approved,omitempty"`
	Comments []DiffComment `json:"comments,omitempty"`
}

// Hunk represents a section of changes
type Hunk struct {
	ID       string `json:"id,omitempty"`
	OldStart int    `json:"oldStart"`
	OldLines int    `json:"oldLines"`
	NewStart int    `json:"newStart"`
	NewLines int    `json:"newLines"`
	Lines    []Line `json:"lines"`
	Approved bool   `json:"approved,omitempty"`
}

// Line represents a single line in a diff
type Line struct {
	Type    LineType `json:"type"`
	Content string   `json:"content"`
	OldNum  int      `json:"oldNum,omitempty"`
	NewNum  int      `json:"newNum,omitempty"`
}

// LineType represents the type of change for a line
type LineType string

const (
	LineTypeContext  LineType = "context"
	LineTypeAddition LineType = "addition"
	LineTypeDeletion LineType = "deletion"
)

// generateHunkID creates a unique identifier for a hunk
func generateHunkID(filePath string, oldStart, newStart int) string {
	input := fmt.Sprintf("%s:%d:%d", filePath, oldStart, newStart)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for shorter IDs
}

// ParseUnifiedDiff parses a unified diff string into FileDiffs
func ParseUnifiedDiff(diffText string) ([]FileDiff, error) {
	diffs := []FileDiff{}

	if diffText == "" {
		return diffs, nil
	}

	lines := strings.Split(diffText, "\n")
	var currentDiff *FileDiff
	var currentHunk *Hunk
	oldLineNum := 0
	newLineNum := 0

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// New file header
		if strings.HasPrefix(line, "diff --git") {
			if currentDiff != nil {
				if currentHunk != nil {
					currentDiff.Hunks = append(currentDiff.Hunks, *currentHunk)
				}
				diffs = append(diffs, *currentDiff)
			}
			currentDiff = &FileDiff{
				Hunks: []Hunk{},
			}
			currentHunk = nil
			continue
		}

		if currentDiff == nil {
			continue
		}

		// Parse file paths
		if strings.HasPrefix(line, "--- ") {
			path := strings.TrimPrefix(line, "--- ")
			if strings.HasPrefix(path, "a/") {
				path = path[2:]
			}
			currentDiff.OldPath = path
			if path == "/dev/null" {
				currentDiff.IsNew = true
			}
			continue
		}

		if strings.HasPrefix(line, "+++ ") {
			path := strings.TrimPrefix(line, "+++ ")
			if strings.HasPrefix(path, "b/") {
				path = path[2:]
			}
			currentDiff.NewPath = path
			if path == "/dev/null" {
				currentDiff.IsDelete = true
			}
			continue
		}

		// Check for binary file
		if strings.Contains(line, "Binary files") {
			currentDiff.IsBinary = true
			continue
		}

		// Skip metadata lines
		if strings.HasPrefix(line, "new file mode") ||
			strings.HasPrefix(line, "deleted file mode") ||
			strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "similarity index") ||
			strings.HasPrefix(line, "rename from") ||
			strings.HasPrefix(line, "rename to") {
			continue
		}

		// Parse hunk header
		if strings.HasPrefix(line, "@@") {
			if currentHunk != nil {
				currentDiff.Hunks = append(currentDiff.Hunks, *currentHunk)
			}

			hunk := parseHunkHeader(line)
			// Generate unique hunk ID
			hunk.ID = generateHunkID(currentDiff.NewPath, hunk.OldStart, hunk.NewStart)
			currentHunk = &hunk
			oldLineNum = hunk.OldStart
			newLineNum = hunk.NewStart
			continue
		}

		if currentHunk == nil {
			continue
		}

		// Parse diff lines
		if len(line) == 0 {
			// Skip trailing empty lines
			if i+1 < len(lines) {
				// Check if next line is part of hunk
				nextLine := lines[i+1]
				if len(nextLine) > 0 && !strings.HasPrefix(nextLine, "diff ") &&
					!strings.HasPrefix(nextLine, "@@") {
					currentHunk.Lines = append(currentHunk.Lines, Line{
						Type:    LineTypeContext,
						Content: "",
						OldNum:  oldLineNum,
						NewNum:  newLineNum,
					})
					oldLineNum++
					newLineNum++
				}
			}
		} else if strings.HasPrefix(line, "+") {
			currentHunk.Lines = append(currentHunk.Lines, Line{
				Type:    LineTypeAddition,
				Content: line[1:],
				NewNum:  newLineNum,
			})
			newLineNum++
		} else if strings.HasPrefix(line, "-") {
			currentHunk.Lines = append(currentHunk.Lines, Line{
				Type:    LineTypeDeletion,
				Content: line[1:],
				OldNum:  oldLineNum,
			})
			oldLineNum++
		} else if strings.HasPrefix(line, " ") {
			currentHunk.Lines = append(currentHunk.Lines, Line{
				Type:    LineTypeContext,
				Content: line[1:],
				OldNum:  oldLineNum,
				NewNum:  newLineNum,
			})
			oldLineNum++
			newLineNum++
		}
	}

	// Don't forget the last diff
	if currentDiff != nil {
		if currentHunk != nil {
			currentDiff.Hunks = append(currentDiff.Hunks, *currentHunk)
		}
		diffs = append(diffs, *currentDiff)
	}

	return diffs, nil
}

// parseHunkHeader parses @@ -a,b +c,d @@ format
func parseHunkHeader(line string) Hunk {
	hunk := Hunk{
		OldStart: 1,
		OldLines: 1,
		NewStart: 1,
		NewLines: 1,
		Lines:    []Line{},
	}

	// Find the @@ markers
	start := strings.Index(line, "@@")
	if start == -1 || start+2 >= len(line) {
		return hunk
	}
	end := strings.Index(line[start+2:], "@@")
	if end == -1 {
		return hunk
	}

	header := strings.TrimSpace(line[start+2 : start+2+end])
	parts := strings.Split(header, " ")

	for _, part := range parts {
		if strings.HasPrefix(part, "-") {
			nums := strings.Split(part[1:], ",")
			if len(nums) >= 1 {
				var n int
				_, _ = parseIntSafe(nums[0], &n)
				hunk.OldStart = n
			}
			if len(nums) >= 2 {
				var n int
				_, _ = parseIntSafe(nums[1], &n)
				hunk.OldLines = n
			}
		} else if strings.HasPrefix(part, "+") {
			nums := strings.Split(part[1:], ",")
			if len(nums) >= 1 {
				var n int
				_, _ = parseIntSafe(nums[0], &n)
				hunk.NewStart = n
			}
			if len(nums) >= 2 {
				var n int
				_, _ = parseIntSafe(nums[1], &n)
				hunk.NewLines = n
			}
		}
	}

	return hunk
}

// parseIntSafe safely parses an integer
func parseIntSafe(s string, result *int) (bool, error) {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	*result = n
	return true, nil
}

// GenerateSideBySide converts a FileDiff to side-by-side format
func GenerateSideBySide(diff FileDiff) []SideBySideLine {
	lines := []SideBySideLine{}

	for _, hunk := range diff.Hunks {
		i := 0
		for i < len(hunk.Lines) {
			line := hunk.Lines[i]

			switch line.Type {
			case LineTypeContext:
				lines = append(lines, SideBySideLine{
					LeftNum:     line.OldNum,
					LeftContent: line.Content,
					RightNum:    line.NewNum,
					RightContent: line.Content,
					Type:        "context",
				})
				i++
			case LineTypeDeletion:
				// Check for modification (deletion followed by addition)
				if i+1 < len(hunk.Lines) && hunk.Lines[i+1].Type == LineTypeAddition {
					lines = append(lines, SideBySideLine{
						LeftNum:      line.OldNum,
						LeftContent:  line.Content,
						RightNum:     hunk.Lines[i+1].NewNum,
						RightContent: hunk.Lines[i+1].Content,
						Type:         "modified",
					})
					i += 2
				} else {
					lines = append(lines, SideBySideLine{
						LeftNum:     line.OldNum,
						LeftContent: line.Content,
						Type:        "deleted",
					})
					i++
				}
			case LineTypeAddition:
				lines = append(lines, SideBySideLine{
					RightNum:     line.NewNum,
					RightContent: line.Content,
					Type:         "added",
				})
				i++
			default:
				i++
			}
		}
	}

	return lines
}

// SideBySideLine represents a line in side-by-side diff view
type SideBySideLine struct {
	LeftNum      int    `json:"leftNum,omitempty"`
	LeftContent  string `json:"leftContent,omitempty"`
	RightNum     int    `json:"rightNum,omitempty"`
	RightContent string `json:"rightContent,omitempty"`
	Type         string `json:"type"` // context, added, deleted, modified
}

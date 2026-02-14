package diff

import (
	"reflect"
	"testing"
)

// Test fixtures for real diff samples
const (
	// Simple modification diff
	simpleDiff = `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,3 @@
 line 1
-line 2
+modified line 2
 line 3
`

	// New file diff
	newFileDiff = `diff --git a/new.txt b/new.txt
new file mode 100644
--- /dev/null
+++ b/new.txt
@@ -0,0 +1,3 @@
+first line
+second line
+third line
`

	// Deleted file diff
	deletedFileDiff = `diff --git a/deleted.txt b/deleted.txt
deleted file mode 100644
--- a/deleted.txt
+++ /dev/null
@@ -1,3 +0,0 @@
-first line
-second line
-third line
`

	// Binary file diff
	binaryFileDiff = `diff --git a/image.png b/image.png
Binary files a/image.png and b/image.png differ
`

	// Multiple hunks diff
	multipleHunksDiff = `diff --git a/multi.txt b/multi.txt
--- a/multi.txt
+++ b/multi.txt
@@ -1,4 +1,4 @@
 line 1
-line 2
+modified line 2
 line 3
 line 4
@@ -10,4 +10,5 @@
 line 10
 line 11
 line 12
+new line 13
 line 14
`

	// Multiple files diff
	multipleFilesDiff = `diff --git a/file1.txt b/file1.txt
--- a/file1.txt
+++ b/file1.txt
@@ -1,2 +1,2 @@
-old content 1
+new content 1
 line 2
diff --git a/file2.txt b/file2.txt
--- a/file2.txt
+++ b/file2.txt
@@ -1,2 +1,2 @@
 line 1
-old content 2
+new content 2
`

	// Complex diff with context
	complexDiff = `diff --git a/complex.go b/complex.go
--- a/complex.go
+++ b/complex.go
@@ -5,10 +5,12 @@ import (
 )

 func main() {
-	fmt.Println("Hello")
+	fmt.Println("Hello, World")
+	fmt.Println("Welcome")

 	x := 10
-	y := 20
+	y := 30
+	z := 40

 	result := x + y
 }
`

	// Empty lines diff
	emptyLinesDiff = `diff --git a/empty.txt b/empty.txt
--- a/empty.txt
+++ b/empty.txt
@@ -1,5 +1,6 @@
 line 1

-line 3
+modified line 3

 line 5
+line 6
`
)

func TestParseUnifiedDiff_EmptyInput(t *testing.T) {
	diffs, err := ParseUnifiedDiff("")
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(diffs) != 0 {
		t.Errorf("Expected 0 diffs for empty input, got %d", len(diffs))
	}
}

func TestParseUnifiedDiff_SimpleDiff(t *testing.T) {
	diffs, err := ParseUnifiedDiff(simpleDiff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(diffs) != 1 {
		t.Fatalf("Expected 1 diff, got %d", len(diffs))
	}

	diff := diffs[0]

	if diff.OldPath != "file.txt" {
		t.Errorf("Expected OldPath 'file.txt', got '%s'", diff.OldPath)
	}

	if diff.NewPath != "file.txt" {
		t.Errorf("Expected NewPath 'file.txt', got '%s'", diff.NewPath)
	}

	if diff.IsNew {
		t.Error("Expected IsNew to be false")
	}

	if diff.IsDelete {
		t.Error("Expected IsDelete to be false")
	}

	if diff.IsBinary {
		t.Error("Expected IsBinary to be false")
	}

	if len(diff.Hunks) != 1 {
		t.Fatalf("Expected 1 hunk, got %d", len(diff.Hunks))
	}

	hunk := diff.Hunks[0]
	if hunk.OldStart != 1 || hunk.OldLines != 3 {
		t.Errorf("Expected hunk old position @1,3, got @%d,%d", hunk.OldStart, hunk.OldLines)
	}

	if hunk.NewStart != 1 || hunk.NewLines != 3 {
		t.Errorf("Expected hunk new position @1,3, got @%d,%d", hunk.NewStart, hunk.NewLines)
	}

	if len(hunk.Lines) != 4 {
		t.Fatalf("Expected 4 lines, got %d", len(hunk.Lines))
	}

	// Check line types
	expectedTypes := []LineType{LineTypeContext, LineTypeDeletion, LineTypeAddition, LineTypeContext}
	for i, expectedType := range expectedTypes {
		if hunk.Lines[i].Type != expectedType {
			t.Errorf("Line %d: expected type %s, got %s", i, expectedType, hunk.Lines[i].Type)
		}
	}

	// Check line content
	if hunk.Lines[0].Content != "line 1" {
		t.Errorf("Expected line 0 content 'line 1', got '%s'", hunk.Lines[0].Content)
	}

	if hunk.Lines[1].Content != "line 2" {
		t.Errorf("Expected line 1 content 'line 2', got '%s'", hunk.Lines[1].Content)
	}

	if hunk.Lines[2].Content != "modified line 2" {
		t.Errorf("Expected line 2 content 'modified line 2', got '%s'", hunk.Lines[2].Content)
	}

	if hunk.Lines[3].Content != "line 3" {
		t.Errorf("Expected line 3 content 'line 3', got '%s'", hunk.Lines[3].Content)
	}
}

func TestParseUnifiedDiff_NewFile(t *testing.T) {
	diffs, err := ParseUnifiedDiff(newFileDiff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(diffs) != 1 {
		t.Fatalf("Expected 1 diff, got %d", len(diffs))
	}

	diff := diffs[0]

	if !diff.IsNew {
		t.Error("Expected IsNew to be true")
	}

	if diff.OldPath != "/dev/null" {
		t.Errorf("Expected OldPath '/dev/null', got '%s'", diff.OldPath)
	}

	if diff.NewPath != "new.txt" {
		t.Errorf("Expected NewPath 'new.txt', got '%s'", diff.NewPath)
	}

	if len(diff.Hunks) != 1 {
		t.Fatalf("Expected 1 hunk, got %d", len(diff.Hunks))
	}

	hunk := diff.Hunks[0]
	if len(hunk.Lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(hunk.Lines))
	}

	// All lines should be additions
	for i, line := range hunk.Lines {
		if line.Type != LineTypeAddition {
			t.Errorf("Line %d: expected addition, got %s", i, line.Type)
		}
	}
}

func TestParseUnifiedDiff_DeletedFile(t *testing.T) {
	diffs, err := ParseUnifiedDiff(deletedFileDiff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(diffs) != 1 {
		t.Fatalf("Expected 1 diff, got %d", len(diffs))
	}

	diff := diffs[0]

	if !diff.IsDelete {
		t.Error("Expected IsDelete to be true")
	}

	if diff.OldPath != "deleted.txt" {
		t.Errorf("Expected OldPath 'deleted.txt', got '%s'", diff.OldPath)
	}

	if diff.NewPath != "/dev/null" {
		t.Errorf("Expected NewPath '/dev/null', got '%s'", diff.NewPath)
	}

	if len(diff.Hunks) != 1 {
		t.Fatalf("Expected 1 hunk, got %d", len(diff.Hunks))
	}

	hunk := diff.Hunks[0]
	if len(hunk.Lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(hunk.Lines))
	}

	// All lines should be deletions
	for i, line := range hunk.Lines {
		if line.Type != LineTypeDeletion {
			t.Errorf("Line %d: expected deletion, got %s", i, line.Type)
		}
	}
}

func TestParseUnifiedDiff_BinaryFile(t *testing.T) {
	diffs, err := ParseUnifiedDiff(binaryFileDiff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(diffs) != 1 {
		t.Fatalf("Expected 1 diff, got %d", len(diffs))
	}

	diff := diffs[0]

	if !diff.IsBinary {
		t.Error("Expected IsBinary to be true")
	}

	if diff.OldPath != "" {
		t.Errorf("Expected empty OldPath for binary before path parsing, got '%s'", diff.OldPath)
	}
}

func TestParseUnifiedDiff_MultipleHunks(t *testing.T) {
	diffs, err := ParseUnifiedDiff(multipleHunksDiff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(diffs) != 1 {
		t.Fatalf("Expected 1 diff, got %d", len(diffs))
	}

	diff := diffs[0]

	if len(diff.Hunks) != 2 {
		t.Fatalf("Expected 2 hunks, got %d", len(diff.Hunks))
	}

	// Check first hunk
	hunk1 := diff.Hunks[0]
	if hunk1.OldStart != 1 || hunk1.OldLines != 4 {
		t.Errorf("Hunk 1: expected @1,4, got @%d,%d", hunk1.OldStart, hunk1.OldLines)
	}

	// Check second hunk
	hunk2 := diff.Hunks[1]
	if hunk2.OldStart != 10 || hunk2.OldLines != 4 {
		t.Errorf("Hunk 2: expected @10,4, got @%d,%d", hunk2.OldStart, hunk2.OldLines)
	}

	if hunk2.NewStart != 10 || hunk2.NewLines != 5 {
		t.Errorf("Hunk 2: expected new @10,5, got @%d,%d", hunk2.NewStart, hunk2.NewLines)
	}
}

func TestParseUnifiedDiff_MultipleFiles(t *testing.T) {
	diffs, err := ParseUnifiedDiff(multipleFilesDiff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(diffs) != 2 {
		t.Fatalf("Expected 2 diffs, got %d", len(diffs))
	}

	// Check first file
	if diffs[0].OldPath != "file1.txt" {
		t.Errorf("Diff 1: expected OldPath 'file1.txt', got '%s'", diffs[0].OldPath)
	}

	// Check second file
	if diffs[1].OldPath != "file2.txt" {
		t.Errorf("Diff 2: expected OldPath 'file2.txt', got '%s'", diffs[1].OldPath)
	}

	// Each file should have 1 hunk
	for i, diff := range diffs {
		if len(diff.Hunks) != 1 {
			t.Errorf("Diff %d: expected 1 hunk, got %d", i, len(diff.Hunks))
		}
	}
}

func TestParseUnifiedDiff_EmptyLines(t *testing.T) {
	diffs, err := ParseUnifiedDiff(emptyLinesDiff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(diffs) != 1 {
		t.Fatalf("Expected 1 diff, got %d", len(diffs))
	}

	diff := diffs[0]
	if len(diff.Hunks) != 1 {
		t.Fatalf("Expected 1 hunk, got %d", len(diff.Hunks))
	}

	hunk := diff.Hunks[0]

	// Check that empty lines are properly handled
	hasEmptyLine := false
	for _, line := range hunk.Lines {
		if line.Content == "" && line.Type == LineTypeContext {
			hasEmptyLine = true
			break
		}
	}

	if !hasEmptyLine {
		t.Error("Expected to find empty context line")
	}
}

func TestParseUnifiedDiff_LineNumbers(t *testing.T) {
	diffs, err := ParseUnifiedDiff(simpleDiff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(diffs) != 1 {
		t.Fatalf("Expected 1 diff, got %d", len(diffs))
	}

	hunk := diffs[0].Hunks[0]

	// Context line should have both old and new numbers
	if hunk.Lines[0].OldNum != 1 || hunk.Lines[0].NewNum != 1 {
		t.Errorf("Context line: expected nums (1,1), got (%d,%d)",
			hunk.Lines[0].OldNum, hunk.Lines[0].NewNum)
	}

	// Deletion line should have old number only
	if hunk.Lines[1].OldNum != 2 {
		t.Errorf("Deletion line: expected old num 2, got %d", hunk.Lines[1].OldNum)
	}

	// Addition line should have new number only
	if hunk.Lines[2].NewNum != 2 {
		t.Errorf("Addition line: expected new num 2, got %d", hunk.Lines[2].NewNum)
	}
}

func TestParseHunkHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Hunk
	}{
		{
			name:  "simple hunk",
			input: "@@ -1,3 +1,3 @@",
			expected: Hunk{
				OldStart: 1,
				OldLines: 3,
				NewStart: 1,
				NewLines: 3,
				Lines:    []Line{},
			},
		},
		{
			name:  "different sizes",
			input: "@@ -10,5 +12,7 @@",
			expected: Hunk{
				OldStart: 10,
				OldLines: 5,
				NewStart: 12,
				NewLines: 7,
				Lines:    []Line{},
			},
		},
		{
			name:  "single line old",
			input: "@@ -5 +5,3 @@",
			expected: Hunk{
				OldStart: 5,
				OldLines: 1, // defaults to 1 when not specified
				NewStart: 5,
				NewLines: 3,
				Lines:    []Line{},
			},
		},
		{
			name:  "with context",
			input: "@@ -1,4 +1,5 @@ function myFunc() {",
			expected: Hunk{
				OldStart: 1,
				OldLines: 4,
				NewStart: 1,
				NewLines: 5,
				Lines:    []Line{},
			},
		},
		{
			name:  "zero lines (new file hunk)",
			input: "@@ -0,0 +1,3 @@",
			expected: Hunk{
				OldStart: 0,
				OldLines: 0,
				NewStart: 1,
				NewLines: 3,
				Lines:    []Line{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHunkHeader(tt.input)

			if result.OldStart != tt.expected.OldStart {
				t.Errorf("OldStart: expected %d, got %d", tt.expected.OldStart, result.OldStart)
			}

			if result.OldLines != tt.expected.OldLines {
				t.Errorf("OldLines: expected %d, got %d", tt.expected.OldLines, result.OldLines)
			}

			if result.NewStart != tt.expected.NewStart {
				t.Errorf("NewStart: expected %d, got %d", tt.expected.NewStart, result.NewStart)
			}

			if result.NewLines != tt.expected.NewLines {
				t.Errorf("NewLines: expected %d, got %d", tt.expected.NewLines, result.NewLines)
			}
		})
	}
}

func TestParseHunkHeader_Malformed(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"no markers", "some random text"},
		{"incomplete", "@@ -1,3"},
		{"empty", ""},
		{"no numbers", "@@ @@"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHunkHeader(tt.input)
			// Should return default hunk without panicking
			if result.Lines == nil {
				t.Error("parseHunkHeader should initialize Lines slice")
			}
		})
	}
}

func TestGenerateSideBySide_Context(t *testing.T) {
	diff := FileDiff{
		Hunks: []Hunk{
			{
				Lines: []Line{
					{Type: LineTypeContext, Content: "unchanged line", OldNum: 1, NewNum: 1},
				},
			},
		},
	}

	result := GenerateSideBySide(diff)

	if len(result) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(result))
	}

	line := result[0]
	if line.Type != "context" {
		t.Errorf("Expected type 'context', got '%s'", line.Type)
	}

	if line.LeftNum != 1 || line.RightNum != 1 {
		t.Errorf("Expected nums (1,1), got (%d,%d)", line.LeftNum, line.RightNum)
	}

	if line.LeftContent != "unchanged line" || line.RightContent != "unchanged line" {
		t.Errorf("Expected same content on both sides, got '%s' and '%s'",
			line.LeftContent, line.RightContent)
	}
}

func TestGenerateSideBySide_Addition(t *testing.T) {
	diff := FileDiff{
		Hunks: []Hunk{
			{
				Lines: []Line{
					{Type: LineTypeAddition, Content: "new line", NewNum: 2},
				},
			},
		},
	}

	result := GenerateSideBySide(diff)

	if len(result) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(result))
	}

	line := result[0]
	if line.Type != "added" {
		t.Errorf("Expected type 'added', got '%s'", line.Type)
	}

	if line.RightNum != 2 {
		t.Errorf("Expected right num 2, got %d", line.RightNum)
	}

	if line.RightContent != "new line" {
		t.Errorf("Expected right content 'new line', got '%s'", line.RightContent)
	}

	if line.LeftContent != "" {
		t.Errorf("Expected empty left content, got '%s'", line.LeftContent)
	}
}

func TestGenerateSideBySide_Deletion(t *testing.T) {
	diff := FileDiff{
		Hunks: []Hunk{
			{
				Lines: []Line{
					{Type: LineTypeDeletion, Content: "deleted line", OldNum: 3},
				},
			},
		},
	}

	result := GenerateSideBySide(diff)

	if len(result) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(result))
	}

	line := result[0]
	if line.Type != "deleted" {
		t.Errorf("Expected type 'deleted', got '%s'", line.Type)
	}

	if line.LeftNum != 3 {
		t.Errorf("Expected left num 3, got %d", line.LeftNum)
	}

	if line.LeftContent != "deleted line" {
		t.Errorf("Expected left content 'deleted line', got '%s'", line.LeftContent)
	}

	if line.RightContent != "" {
		t.Errorf("Expected empty right content, got '%s'", line.RightContent)
	}
}

func TestGenerateSideBySide_Modification(t *testing.T) {
	diff := FileDiff{
		Hunks: []Hunk{
			{
				Lines: []Line{
					{Type: LineTypeDeletion, Content: "old content", OldNum: 5},
					{Type: LineTypeAddition, Content: "new content", NewNum: 5},
				},
			},
		},
	}

	result := GenerateSideBySide(diff)

	if len(result) != 1 {
		t.Fatalf("Expected 1 line (modification), got %d", len(result))
	}

	line := result[0]
	if line.Type != "modified" {
		t.Errorf("Expected type 'modified', got '%s'", line.Type)
	}

	if line.LeftNum != 5 || line.RightNum != 5 {
		t.Errorf("Expected nums (5,5), got (%d,%d)", line.LeftNum, line.RightNum)
	}

	if line.LeftContent != "old content" {
		t.Errorf("Expected left content 'old content', got '%s'", line.LeftContent)
	}

	if line.RightContent != "new content" {
		t.Errorf("Expected right content 'new content', got '%s'", line.RightContent)
	}
}

func TestGenerateSideBySide_Mixed(t *testing.T) {
	diff := FileDiff{
		Hunks: []Hunk{
			{
				Lines: []Line{
					{Type: LineTypeContext, Content: "line 1", OldNum: 1, NewNum: 1},
					{Type: LineTypeDeletion, Content: "old line 2", OldNum: 2},
					{Type: LineTypeAddition, Content: "new line 2", NewNum: 2},
					{Type: LineTypeContext, Content: "line 3", OldNum: 3, NewNum: 3},
					{Type: LineTypeAddition, Content: "added line 4", NewNum: 4},
					{Type: LineTypeDeletion, Content: "deleted line", OldNum: 4},
				},
			},
		},
	}

	result := GenerateSideBySide(diff)

	if len(result) != 5 {
		t.Fatalf("Expected 5 lines, got %d", len(result))
	}

	expectedTypes := []string{"context", "modified", "context", "added", "deleted"}
	for i, expectedType := range expectedTypes {
		if result[i].Type != expectedType {
			t.Errorf("Line %d: expected type '%s', got '%s'", i, expectedType, result[i].Type)
		}
	}
}

func TestGenerateSideBySide_Empty(t *testing.T) {
	diff := FileDiff{
		Hunks: []Hunk{},
	}

	result := GenerateSideBySide(diff)

	if len(result) != 0 {
		t.Errorf("Expected 0 lines for empty diff, got %d", len(result))
	}
}

func TestGenerateSideBySide_MultipleHunks(t *testing.T) {
	diff := FileDiff{
		Hunks: []Hunk{
			{
				Lines: []Line{
					{Type: LineTypeContext, Content: "line 1", OldNum: 1, NewNum: 1},
					{Type: LineTypeDeletion, Content: "old", OldNum: 2},
					{Type: LineTypeAddition, Content: "new", NewNum: 2},
				},
			},
			{
				Lines: []Line{
					{Type: LineTypeContext, Content: "line 10", OldNum: 10, NewNum: 10},
					{Type: LineTypeAddition, Content: "added", NewNum: 11},
				},
			},
		},
	}

	result := GenerateSideBySide(diff)

	// Should have lines from both hunks
	if len(result) != 4 {
		t.Fatalf("Expected 4 lines total, got %d", len(result))
	}
}

func TestLineType_Constants(t *testing.T) {
	if LineTypeContext != "context" {
		t.Errorf("LineTypeContext should be 'context', got '%s'", LineTypeContext)
	}

	if LineTypeAddition != "addition" {
		t.Errorf("LineTypeAddition should be 'addition', got '%s'", LineTypeAddition)
	}

	if LineTypeDeletion != "deletion" {
		t.Errorf("LineTypeDeletion should be 'deletion', got '%s'", LineTypeDeletion)
	}
}

func TestFileDiff_Structure(t *testing.T) {
	diff := FileDiff{
		OldPath:  "old.txt",
		NewPath:  "new.txt",
		IsNew:    true,
		IsDelete: false,
		IsBinary: false,
		Hunks: []Hunk{
			{
				OldStart: 1,
				OldLines: 5,
				NewStart: 1,
				NewLines: 6,
				Lines:    []Line{},
			},
		},
	}

	if diff.OldPath != "old.txt" {
		t.Error("FileDiff OldPath not set correctly")
	}

	if !diff.IsNew {
		t.Error("FileDiff IsNew not set correctly")
	}

	if len(diff.Hunks) != 1 {
		t.Error("FileDiff Hunks not set correctly")
	}
}

func TestParseIntSafe(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"simple number", "42", 42},
		{"zero", "0", 0},
		{"large number", "12345", 12345},
		{"number with non-digits", "123abc", 123},
		{"non-digit start", "abc", 0},
		{"empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result int
			parseIntSafe(tt.input, &result)

			if result != tt.expected {
				t.Errorf("parseIntSafe(%s) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseUnifiedDiff_ComplexCase(t *testing.T) {
	diffs, err := ParseUnifiedDiff(complexDiff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(diffs) != 1 {
		t.Fatalf("Expected 1 diff, got %d", len(diffs))
	}

	diff := diffs[0]

	if diff.NewPath != "complex.go" {
		t.Errorf("Expected NewPath 'complex.go', got '%s'", diff.NewPath)
	}

	if len(diff.Hunks) != 1 {
		t.Fatalf("Expected 1 hunk, got %d", len(diff.Hunks))
	}

	hunk := diff.Hunks[0]

	// Count different line types
	contextCount := 0
	addCount := 0
	delCount := 0

	for _, line := range hunk.Lines {
		switch line.Type {
		case LineTypeContext:
			contextCount++
		case LineTypeAddition:
			addCount++
		case LineTypeDeletion:
			delCount++
		}
	}

	if contextCount == 0 {
		t.Error("Expected some context lines")
	}

	if addCount == 0 {
		t.Error("Expected some addition lines")
	}

	if delCount == 0 {
		t.Error("Expected some deletion lines")
	}
}

func TestParseUnifiedDiff_PathParsing(t *testing.T) {
	tests := []struct {
		name        string
		diff        string
		expectedOld string
		expectedNew string
	}{
		{
			name: "standard paths with a/ b/ prefix",
			diff: `diff --git a/path/to/file.txt b/path/to/file.txt
--- a/path/to/file.txt
+++ b/path/to/file.txt
@@ -1 +1 @@
-old
+new
`,
			expectedOld: "path/to/file.txt",
			expectedNew: "path/to/file.txt",
		},
		{
			name: "paths without prefix",
			diff: `diff --git a/file.txt b/file.txt
--- file.txt
+++ file.txt
@@ -1 +1 @@
-old
+new
`,
			expectedOld: "file.txt",
			expectedNew: "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs, err := ParseUnifiedDiff(tt.diff)
			if err != nil {
				t.Fatalf("ParseUnifiedDiff() error = %v", err)
			}

			if len(diffs) != 1 {
				t.Fatalf("Expected 1 diff, got %d", len(diffs))
			}

			if diffs[0].OldPath != tt.expectedOld {
				t.Errorf("Expected OldPath '%s', got '%s'", tt.expectedOld, diffs[0].OldPath)
			}

			if diffs[0].NewPath != tt.expectedNew {
				t.Errorf("Expected NewPath '%s', got '%s'", tt.expectedNew, diffs[0].NewPath)
			}
		})
	}
}

func TestSideBySideLine_Structure(t *testing.T) {
	line := SideBySideLine{
		LeftNum:      10,
		LeftContent:  "left",
		RightNum:     20,
		RightContent: "right",
		Type:         "modified",
	}

	if line.LeftNum != 10 {
		t.Error("LeftNum not set correctly")
	}

	if line.Type != "modified" {
		t.Error("Type not set correctly")
	}
}

func TestGenerateSideBySide_ConsecutiveDeletions(t *testing.T) {
	diff := FileDiff{
		Hunks: []Hunk{
			{
				Lines: []Line{
					{Type: LineTypeDeletion, Content: "deleted 1", OldNum: 1},
					{Type: LineTypeDeletion, Content: "deleted 2", OldNum: 2},
					{Type: LineTypeAddition, Content: "added", NewNum: 1},
				},
			},
		},
	}

	result := GenerateSideBySide(diff)

	// Current behavior: pairs the last deletion with the addition as modification
	// First deletion stands alone, second deletion + addition = modified
	if len(result) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(result))
	}

	if result[0].Type != "deleted" {
		t.Errorf("Expected first line to be 'deleted', got '%s'", result[0].Type)
	}

	if result[1].Type != "modified" {
		t.Errorf("Expected second line to be 'modified', got '%s'", result[1].Type)
	}
}

func TestParseUnifiedDiff_RealWorldGo(t *testing.T) {
	goFileDiff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,12 +1,15 @@
 package main

 import (
-	"fmt"
+	"fmt"
+	"os"
 )

 func main() {
-	name := "World"
-	fmt.Printf("Hello, %s!\n", name)
+	if len(os.Args) > 1 {
+		name := os.Args[1]
+		fmt.Printf("Hello, %s!\n", name)
+	}
 }
`

	diffs, err := ParseUnifiedDiff(goFileDiff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(diffs) != 1 {
		t.Fatalf("Expected 1 diff, got %d", len(diffs))
	}

	diff := diffs[0]

	if diff.NewPath != "main.go" {
		t.Errorf("Expected NewPath 'main.go', got '%s'", diff.NewPath)
	}

	if diff.IsNew || diff.IsDelete || diff.IsBinary {
		t.Error("Regular file modification should not be marked as new/delete/binary")
	}

	if len(diff.Hunks) != 1 {
		t.Fatalf("Expected 1 hunk, got %d", len(diff.Hunks))
	}
}

func TestHunk_DeepEqual(t *testing.T) {
	hunk1 := Hunk{
		OldStart: 1,
		OldLines: 5,
		NewStart: 1,
		NewLines: 5,
		Lines: []Line{
			{Type: LineTypeContext, Content: "test", OldNum: 1, NewNum: 1},
		},
	}

	hunk2 := Hunk{
		OldStart: 1,
		OldLines: 5,
		NewStart: 1,
		NewLines: 5,
		Lines: []Line{
			{Type: LineTypeContext, Content: "test", OldNum: 1, NewNum: 1},
		},
	}

	if !reflect.DeepEqual(hunk1, hunk2) {
		t.Error("Identical hunks should be deeply equal")
	}

	hunk3 := hunk1
	hunk3.OldStart = 2

	if reflect.DeepEqual(hunk1, hunk3) {
		t.Error("Different hunks should not be deeply equal")
	}
}

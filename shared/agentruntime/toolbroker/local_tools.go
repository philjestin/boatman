package toolbroker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

const (
	defaultReadLimit   = 1 << 20
	defaultGrepMatches = 200
	defaultBashTimeout = 30 * time.Second
)

// ReadTool reads a UTF-8-ish file from the workspace root.
type ReadTool struct{}

func (ReadTool) Name() string { return "Read" }

func (ReadTool) Ref() agentruntime.ToolRef {
	return agentruntime.ToolRef{
		Name:        "Read",
		Kind:        "local",
		Description: "Read a file from the workspace.",
		Schema: schema(map[string]any{
			"file_path": textProperty("Workspace-relative or in-root absolute file path."),
			"limit":     numberProperty("Optional maximum bytes to return."),
			"offset":    numberProperty("Optional byte offset to start reading from."),
		}, "file_path"),
	}
}

func (ReadTool) Invoke(_ context.Context, inv Invocation) (Result, error) {
	input, err := decodeInput[struct {
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
		Limit    int    `json:"limit"`
		Offset   int64  `json:"offset"`
	}](inv.Input)
	if err != nil {
		return Result{}, err
	}
	path := input.FilePath
	if path == "" {
		path = input.Path
	}
	fullPath, relPath, err := resolvePath(inv.WorkDir, path)
	if err != nil {
		return Result{}, err
	}
	limit := input.Limit
	if limit <= 0 || limit > defaultReadLimit {
		limit = defaultReadLimit
	}
	file, err := os.Open(fullPath)
	if err != nil {
		return Result{}, err
	}
	defer file.Close()
	if input.Offset > 0 {
		if _, err := file.Seek(input.Offset, 0); err != nil {
			return Result{}, err
		}
	}
	buf := make([]byte, limit+1)
	n, err := file.Read(buf)
	if err != nil && n == 0 {
		return Result{}, err
	}
	truncated := n > limit
	if n > limit {
		n = limit
	}
	return jsonResult(map[string]any{
		"path":      relPath,
		"content":   string(buf[:n]),
		"truncated": truncated,
	})
}

// WriteTool writes complete file contents under the workspace root.
type WriteTool struct{}

func (WriteTool) Name() string { return "Write" }

func (WriteTool) Ref() agentruntime.ToolRef {
	return agentruntime.ToolRef{
		Name:        "Write",
		Kind:        "local",
		Description: "Write complete file contents under the workspace.",
		Schema: schema(map[string]any{
			"file_path": textProperty("Workspace-relative or in-root absolute file path."),
			"content":   textProperty("Complete file content to write."),
		}, "file_path", "content"),
	}
}

func (WriteTool) Invoke(_ context.Context, inv Invocation) (Result, error) {
	if inv.ApprovalPolicy == agentruntime.ApprovalSuggest {
		return Result{}, fmt.Errorf("Write requires auto_edit or full_auto approval")
	}
	input, err := decodeInput[struct {
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
		Content  string `json:"content"`
	}](inv.Input)
	if err != nil {
		return Result{}, err
	}
	path := input.FilePath
	if path == "" {
		path = input.Path
	}
	fullPath, relPath, err := resolvePath(inv.WorkDir, path)
	if err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return Result{}, err
	}
	if err := os.WriteFile(fullPath, []byte(input.Content), 0644); err != nil {
		return Result{}, err
	}
	return jsonResult(map[string]any{
		"path":          relPath,
		"bytes_written": len(input.Content),
	})
}

// EditTool replaces text in an existing file under the workspace root.
type EditTool struct{}

func (EditTool) Name() string { return "Edit" }

func (EditTool) Ref() agentruntime.ToolRef {
	return agentruntime.ToolRef{
		Name:        "Edit",
		Kind:        "local",
		Description: "Replace text in an existing workspace file.",
		Schema: schema(map[string]any{
			"file_path":   textProperty("Workspace-relative or in-root absolute file path."),
			"old_string":  textProperty("Existing text to replace."),
			"new_string":  textProperty("Replacement text."),
			"replace_all": boolProperty("Replace every occurrence instead of exactly one."),
		}, "file_path", "old_string", "new_string"),
	}
}

func (EditTool) Invoke(_ context.Context, inv Invocation) (Result, error) {
	if inv.ApprovalPolicy == agentruntime.ApprovalSuggest {
		return Result{}, fmt.Errorf("Edit requires auto_edit or full_auto approval")
	}
	input, err := decodeInput[struct {
		FilePath   string `json:"file_path"`
		Path       string `json:"path"`
		OldString  string `json:"old_string"`
		NewString  string `json:"new_string"`
		ReplaceAll bool   `json:"replace_all"`
	}](inv.Input)
	if err != nil {
		return Result{}, err
	}
	path := input.FilePath
	if path == "" {
		path = input.Path
	}
	fullPath, relPath, err := resolvePath(inv.WorkDir, path)
	if err != nil {
		return Result{}, err
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return Result{}, err
	}
	content := string(data)
	count := strings.Count(content, input.OldString)
	if count == 0 {
		return Result{}, fmt.Errorf("old_string not found in %s", relPath)
	}
	if !input.ReplaceAll && count != 1 {
		return Result{}, fmt.Errorf("old_string appears %d times in %s; set replace_all to true or use a more specific string", count, relPath)
	}
	replacements := 1
	updated := strings.Replace(content, input.OldString, input.NewString, 1)
	if input.ReplaceAll {
		replacements = count
		updated = strings.ReplaceAll(content, input.OldString, input.NewString)
	}
	if err := os.WriteFile(fullPath, []byte(updated), 0644); err != nil {
		return Result{}, err
	}
	return jsonResult(map[string]any{
		"path":         relPath,
		"replacements": replacements,
	})
}

// BashTool executes a shell command under the workspace root.
type BashTool struct{}

func (BashTool) Name() string { return "Bash" }

func (BashTool) Ref() agentruntime.ToolRef {
	return agentruntime.ToolRef{
		Name:        "Bash",
		Kind:        "local",
		Description: "Run a shell command in the workspace. Requires full_auto approval.",
		Schema: schema(map[string]any{
			"command":    textProperty("Shell command to run."),
			"timeout_ms": numberProperty("Optional timeout in milliseconds."),
		}, "command"),
	}
}

func (BashTool) Invoke(ctx context.Context, inv Invocation) (Result, error) {
	if inv.ApprovalPolicy != agentruntime.ApprovalFullAuto {
		return Result{}, fmt.Errorf("Bash requires full_auto approval")
	}
	input, err := decodeInput[struct {
		Command   string `json:"command"`
		TimeoutMS int    `json:"timeout_ms"`
	}](inv.Input)
	if err != nil {
		return Result{}, err
	}
	if strings.TrimSpace(input.Command) == "" {
		return Result{}, fmt.Errorf("command is required")
	}
	root, err := workspaceRoot(inv.WorkDir)
	if err != nil {
		return Result{}, err
	}
	timeout := defaultBashTimeout
	if input.TimeoutMS > 0 {
		timeout = time.Duration(input.TimeoutMS) * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", input.Command)
	cmd.Dir = root
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		err = fmt.Errorf("command timed out after %s", timeout)
	}
	result, marshalErr := jsonResult(map[string]any{
		"stdout":     stdout.String(),
		"stderr":     stderr.String(),
		"exit_error": exitError(err),
	})
	if marshalErr != nil {
		return Result{}, marshalErr
	}
	return result, err
}

// GrepTool searches text files under the workspace root.
type GrepTool struct{}

func (GrepTool) Name() string { return "Grep" }

func (GrepTool) Ref() agentruntime.ToolRef {
	return agentruntime.ToolRef{
		Name:        "Grep",
		Kind:        "local",
		Description: "Search workspace text files with a regular expression.",
		Schema: schema(map[string]any{
			"pattern":     textProperty("Regular expression to search for."),
			"path":        textProperty("Optional workspace-relative directory or file to search."),
			"glob":        textProperty("Optional glob filter, such as **/*.go."),
			"ignore_case": boolProperty("Whether matching is case-insensitive."),
			"max_matches": numberProperty("Maximum matches to return."),
		}, "pattern"),
	}
}

func (GrepTool) Invoke(_ context.Context, inv Invocation) (Result, error) {
	input, err := decodeInput[struct {
		Pattern    string `json:"pattern"`
		Path       string `json:"path"`
		Glob       string `json:"glob"`
		IgnoreCase bool   `json:"ignore_case"`
		MaxMatches int    `json:"max_matches"`
	}](inv.Input)
	if err != nil {
		return Result{}, err
	}
	pattern := input.Pattern
	if input.IgnoreCase {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return Result{}, err
	}
	root, searchRoot, err := searchRoot(inv.WorkDir, input.Path)
	if err != nil {
		return Result{}, err
	}
	maxMatches := input.MaxMatches
	if maxMatches <= 0 {
		maxMatches = defaultGrepMatches
	}
	var matches []map[string]any
	err = walkTextFiles(searchRoot, func(path string) error {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if input.Glob != "" && !matchGlob(input.Glob, rel) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()
			if re.MatchString(line) {
				matches = append(matches, map[string]any{
					"path": rel,
					"line": lineNo,
					"text": line,
				})
				if len(matches) >= maxMatches {
					return errStopWalk
				}
			}
		}
		return nil
	})
	if err == errStopWalk {
		err = nil
	}
	if err != nil {
		return Result{}, err
	}
	return jsonResult(map[string]any{
		"matches":   matches,
		"truncated": len(matches) >= maxMatches,
	})
}

// GlobTool lists workspace paths matching a glob pattern.
type GlobTool struct{}

func (GlobTool) Name() string { return "Glob" }

func (GlobTool) Ref() agentruntime.ToolRef {
	return agentruntime.ToolRef{
		Name:        "Glob",
		Kind:        "local",
		Description: "List workspace files matching a glob pattern.",
		Schema: schema(map[string]any{
			"pattern": textProperty("Glob pattern, such as **/*.go."),
			"path":    textProperty("Optional workspace-relative directory to search."),
		}, "pattern"),
	}
}

func (GlobTool) Invoke(_ context.Context, inv Invocation) (Result, error) {
	input, err := decodeInput[struct {
		Pattern string `json:"pattern"`
		Path    string `json:"path"`
	}](inv.Input)
	if err != nil {
		return Result{}, err
	}
	if strings.TrimSpace(input.Pattern) == "" {
		return Result{}, fmt.Errorf("pattern is required")
	}
	root, searchRoot, err := searchRoot(inv.WorkDir, input.Path)
	if err != nil {
		return Result{}, err
	}
	var paths []string
	err = filepath.WalkDir(searchRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if shouldSkip(entry) {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if matchGlob(input.Pattern, rel) {
			paths = append(paths, rel)
		}
		return nil
	})
	if err != nil {
		return Result{}, err
	}
	sort.Strings(paths)
	return jsonResult(map[string]any{"paths": paths})
}

var errStopWalk = fmt.Errorf("stop walk")

func jsonResult(value any) (Result, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return Result{}, err
	}
	return Result{Output: data}, nil
}

func workspaceRoot(workDir string) (string, error) {
	root := normalizeToolPath(workDir)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.Clean(root), nil
}

func resolvePath(workDir, candidate string) (string, string, error) {
	candidate = normalizeToolPath(candidate)
	if candidate == "" {
		return "", "", fmt.Errorf("file_path is required")
	}
	root, err := workspaceRoot(workDir)
	if err != nil {
		return "", "", err
	}
	fullPath := candidate
	if !filepath.IsAbs(fullPath) {
		fullPath = filepath.Join(root, fullPath)
	}
	fullPath, err = filepath.Abs(fullPath)
	if err != nil {
		return "", "", err
	}
	fullPath = filepath.Clean(fullPath)
	rel, err := filepath.Rel(root, fullPath)
	if err != nil {
		return "", "", err
	}
	if rel == "." {
		return fullPath, rel, nil
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", "", fmt.Errorf("path %q escapes workspace root", candidate)
	}
	return fullPath, filepath.ToSlash(rel), nil
}

func searchRoot(workDir, path string) (string, string, error) {
	root, err := workspaceRoot(workDir)
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(path) == "" {
		return root, root, nil
	}
	fullPath, _, err := resolvePath(workDir, path)
	if err != nil {
		return "", "", err
	}
	return root, fullPath, nil
}

func walkTextFiles(root string, visit func(path string) error) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if shouldSkip(entry) {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		if !looksText(path) {
			return nil
		}
		return visit(path)
	})
}

func shouldSkip(entry os.DirEntry) bool {
	name := entry.Name()
	return entry.IsDir() && (name == ".git" || name == "node_modules" || name == "dist" || name == "build")
}

func looksText(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	buf := make([]byte, 1024)
	n, _ := file.Read(buf)
	return !bytes.Contains(buf[:n], []byte{0})
}

func matchGlob(pattern, rel string) bool {
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))
	rel = filepath.ToSlash(rel)
	if pattern == "" {
		return true
	}
	if strings.HasPrefix(pattern, "**/") && matchGlob(strings.TrimPrefix(pattern, "**/"), rel) {
		return true
	}
	if ok, _ := filepath.Match(pattern, rel); ok {
		return true
	}
	regex := regexp.QuoteMeta(pattern)
	regex = strings.ReplaceAll(regex, `\*\*`, `.*`)
	regex = strings.ReplaceAll(regex, `\*`, `[^/]*`)
	regex = strings.ReplaceAll(regex, `\?`, `[^/]`)
	ok, _ := regexp.MatchString("^"+regex+"$", rel)
	return ok
}

func exitError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

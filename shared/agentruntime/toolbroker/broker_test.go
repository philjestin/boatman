package toolbroker

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

func TestLocalRefsExposeSchemas(t *testing.T) {
	refs := ExecutorRefs()
	names := make([]string, 0, len(refs))
	for _, ref := range refs {
		names = append(names, ref.Name)
		if ref.Kind != "local" {
			t.Fatalf("%s kind = %q, want local", ref.Name, ref.Kind)
		}
		if !json.Valid(ref.Schema) {
			t.Fatalf("%s schema is invalid JSON: %s", ref.Name, ref.Schema)
		}
	}

	want := "Read,Write,Edit,Bash,Grep,Glob"
	if strings.Join(names, ",") != want {
		t.Fatalf("tool order = %v, want %s", names, want)
	}
}

func TestBrokerInvokeReadWriteEdit(t *testing.T) {
	root := t.TempDir()
	broker := NewLocal()

	write := invocation("Write", root, agentruntime.ApprovalAutoEdit, map[string]any{
		"file_path": "dir/file.txt",
		"content":   "hello world",
	})
	if _, err := broker.Invoke(context.Background(), write); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	edit := invocation("Edit", root, agentruntime.ApprovalAutoEdit, map[string]any{
		"file_path":  "dir/file.txt",
		"old_string": "world",
		"new_string": "broker",
	})
	if _, err := broker.Invoke(context.Background(), edit); err != nil {
		t.Fatalf("Edit failed: %v", err)
	}

	read := invocation("Read", root, agentruntime.ApprovalSuggest, map[string]any{
		"file_path": "dir/file.txt",
	})
	result, err := broker.Invoke(context.Background(), read)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(result.Output, &output); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if output["content"] != "hello broker" {
		t.Fatalf("content = %#v, want hello broker", output["content"])
	}
}

func TestBrokerRejectsPathEscape(t *testing.T) {
	root := t.TempDir()
	broker := NewLocal()

	_, err := broker.Invoke(context.Background(), invocation("Read", root, agentruntime.ApprovalSuggest, map[string]any{
		"file_path": "../outside.txt",
	}))
	if err == nil {
		t.Fatal("Read should reject paths outside the workspace")
	}
}

func TestGrepAndGlob(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "main.go"), "package main\nfunc main() {}\n")
	mustWrite(t, filepath.Join(root, "README.md"), "main docs\n")

	broker := NewLocal()
	globResult, err := broker.Invoke(context.Background(), invocation("Glob", root, agentruntime.ApprovalSuggest, map[string]any{
		"pattern": "**/*.go",
	}))
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}
	var globOutput struct {
		Paths []string `json:"paths"`
	}
	if err := json.Unmarshal(globResult.Output, &globOutput); err != nil {
		t.Fatalf("decode glob: %v", err)
	}
	if len(globOutput.Paths) != 1 || globOutput.Paths[0] != "main.go" {
		t.Fatalf("glob paths = %#v, want main.go", globOutput.Paths)
	}

	grepResult, err := broker.Invoke(context.Background(), invocation("Grep", root, agentruntime.ApprovalSuggest, map[string]any{
		"pattern": "func main",
		"glob":    "**/*.go",
	}))
	if err != nil {
		t.Fatalf("Grep failed: %v", err)
	}
	var grepOutput struct {
		Matches []struct {
			Path string `json:"path"`
			Line int    `json:"line"`
			Text string `json:"text"`
		} `json:"matches"`
	}
	if err := json.Unmarshal(grepResult.Output, &grepOutput); err != nil {
		t.Fatalf("decode grep: %v", err)
	}
	if len(grepOutput.Matches) != 1 || grepOutput.Matches[0].Path != "main.go" || grepOutput.Matches[0].Line != 2 {
		t.Fatalf("grep matches = %#v, want main.go line 2", grepOutput.Matches)
	}
}

func TestApprovalGates(t *testing.T) {
	root := t.TempDir()
	broker := NewLocal()

	_, err := broker.Invoke(context.Background(), invocation("Write", root, agentruntime.ApprovalSuggest, map[string]any{
		"file_path": "file.txt",
		"content":   "nope",
	}))
	if err == nil {
		t.Fatal("Write should require edit approval")
	}

	if runtime.GOOS == "windows" {
		t.Skip("Bash tool uses POSIX shell in tests")
	}
	_, err = broker.Invoke(context.Background(), invocation("Bash", root, agentruntime.ApprovalAutoEdit, map[string]any{
		"command": "echo no",
	}))
	if err == nil {
		t.Fatal("Bash should require full_auto approval")
	}

	result, err := broker.Invoke(context.Background(), invocation("Bash", root, agentruntime.ApprovalFullAuto, map[string]any{
		"command": "printf ok",
	}))
	if err != nil {
		t.Fatalf("Bash full_auto failed: %v", err)
	}
	var output map[string]any
	if err := json.Unmarshal(result.Output, &output); err != nil {
		t.Fatalf("decode bash: %v", err)
	}
	if output["stdout"] != "ok" {
		t.Fatalf("stdout = %#v, want ok", output["stdout"])
	}
}

func invocation(name, workDir string, approval agentruntime.ApprovalPolicy, input map[string]any) Invocation {
	raw, _ := json.Marshal(input)
	return Invocation{
		ID:             "tool-1",
		Name:           name,
		WorkDir:        workDir,
		Input:          raw,
		ApprovalPolicy: approval,
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

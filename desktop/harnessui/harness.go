package harnessui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/philjestin/boatman-ecosystem/harness/scaffold"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ScaffoldRequest describes the parameters for generating a new harness project.
type ScaffoldRequest struct {
	ProjectName         string `json:"projectName"`
	OutputDir           string `json:"outputDir"`
	Provider            string `json:"provider"`
	ProjectLang         string `json:"projectLang"`
	IncludePlanner      bool   `json:"includePlanner"`
	IncludeTester       bool   `json:"includeTester"`
	IncludeCostTracking bool   `json:"includeCostTracking"`
	MaxIterations       int    `json:"maxIterations"`
	ReviewCriteria      string `json:"reviewCriteria"`
}

// ScaffoldResponse describes the result of a scaffold generation.
type ScaffoldResponse struct {
	OutputDir    string   `json:"outputDir"`
	FilesCreated []string `json:"filesCreated"`
}

// HarnessInfo describes a discovered harness project.
type HarnessInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	HasGoMod bool   `json:"hasGoMod"`
	HasMain  bool   `json:"hasMain"`
}

// RunRequest describes the parameters for running a harness.
type RunRequest struct {
	HarnessPath     string            `json:"harnessPath"`
	WorkDir         string            `json:"workDir"`
	TaskTitle       string            `json:"taskTitle"`
	TaskDescription string            `json:"taskDescription"`
	EnvVars         map[string]string `json:"envVars"`
}

// GenerateScaffold creates a new harness project using the scaffold package.
// If OutputDir is empty, it defaults to ~/.boatman/harnesses/{last-segment}/.
func GenerateScaffold(req ScaffoldRequest) (*ScaffoldResponse, error) {
	outputDir := req.OutputDir
	if outputDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		lastSegment := filepath.Base(req.ProjectName)
		outputDir = filepath.Join(homeDir, ".boatman", "harnesses", lastSegment)
	}

	cfg := scaffold.ScaffoldConfig{
		ProjectName:         req.ProjectName,
		OutputDir:           outputDir,
		Provider:            scaffold.LLMProvider(req.Provider),
		ProjectLang:         scaffold.ProjectLanguage(req.ProjectLang),
		IncludePlanner:      req.IncludePlanner,
		IncludeTester:       req.IncludeTester,
		IncludeCostTracking: req.IncludeCostTracking,
		MaxIterations:       req.MaxIterations,
		ReviewCriteria:      req.ReviewCriteria,
	}

	result, err := scaffold.Generate(cfg)
	if err != nil {
		return nil, fmt.Errorf("scaffold generation failed: %w", err)
	}

	return &ScaffoldResponse{
		OutputDir:    result.OutputDir,
		FilesCreated: result.FilesCreated,
	}, nil
}

// ListHarnesses scans ~/.boatman/harnesses/ and returns info about each project.
func ListHarnesses() ([]HarnessInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	harnessDir := filepath.Join(homeDir, ".boatman", "harnesses")
	entries, err := os.ReadDir(harnessDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []HarnessInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read harnesses directory: %w", err)
	}

	var harnesses []HarnessInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(harnessDir, entry.Name())
		info := HarnessInfo{
			Name: entry.Name(),
			Path: dirPath,
		}

		if _, err := os.Stat(filepath.Join(dirPath, "go.mod")); err == nil {
			info.HasGoMod = true
		}
		if _, err := os.Stat(filepath.Join(dirPath, "main.go")); err == nil {
			info.HasMain = true
		}

		harnesses = append(harnesses, info)
	}

	if harnesses == nil {
		harnesses = []HarnessInfo{}
	}

	return harnesses, nil
}

// RunHarness executes `go run .` in the harness directory as a subprocess,
// streaming stdout/stderr via Wails events.
func RunHarness(ctx context.Context, wailsCtx context.Context, runID string, req RunRequest) error {
	cmd := exec.CommandContext(ctx, "go", "run", ".")
	cmd.Dir = req.HarnessPath

	// Build environment
	cmd.Env = os.Environ()
	if req.TaskTitle != "" {
		cmd.Env = append(cmd.Env, "HARNESS_TASK_TITLE="+req.TaskTitle)
	}
	if req.TaskDescription != "" {
		cmd.Env = append(cmd.Env, "HARNESS_TASK_DESCRIPTION="+req.TaskDescription)
	}
	if req.WorkDir != "" {
		cmd.Env = append(cmd.Env, "HARNESS_WORK_DIR="+req.WorkDir)
	}
	for k, v := range req.EnvVars {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start harness: %w", err)
	}

	runtime.EventsEmit(wailsCtx, "harness:started", map[string]any{
		"runId": runID,
	})

	// Stream stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			runtime.EventsEmit(wailsCtx, "harness:output", map[string]any{
				"runId":  runID,
				"line":   scanner.Text(),
				"stream": "stdout",
			})
		}
	}()

	// Stream stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			runtime.EventsEmit(wailsCtx, "harness:output", map[string]any{
				"runId":  runID,
				"line":   scanner.Text(),
				"stream": "stderr",
			})
		}
	}()

	return cmd.Wait()
}

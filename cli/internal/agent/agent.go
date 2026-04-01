// Package agent orchestrates the complete workflow:
// fetch ticket → create worktree → validate → execute → test → review → verify → refactor → PR
package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/philjestin/boatmanmode/internal/brain"
	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/contextpin"
	"github.com/philjestin/boatmanmode/internal/coordinator"
	"github.com/philjestin/boatmanmode/internal/cost"
	"github.com/philjestin/boatmanmode/internal/diffverify"
	"github.com/philjestin/boatmanmode/internal/events"
	"github.com/philjestin/boatmanmode/internal/executor"
	"github.com/philjestin/boatmanmode/internal/github"
	"github.com/philjestin/boatmanmode/internal/handoff"
	"github.com/philjestin/boatmanmode/internal/linear"
	"github.com/philjestin/boatmanmode/internal/planner"
	"github.com/philjestin/boatmanmode/internal/preflight"
	"github.com/philjestin/boatmanmode/internal/scottbott"
	"github.com/philjestin/boatmanmode/internal/task"
	"github.com/philjestin/boatmanmode/internal/testrunner"
	"github.com/philjestin/boatmanmode/internal/worktree"
)

// Agent orchestrates the development workflow.
type Agent struct {
	config       *config.Config
	linearClient *linear.Client
	coordinator  *coordinator.Coordinator

	// PreloadedPlan, if set, is used instead of running the planning agent.
	// This allows triage-generated plans to be reused during execution.
	PreloadedPlan *planner.Plan
}

// WorkResult represents the outcome of the work command.
type WorkResult struct {
	PRCreated    bool
	PRURL        string
	Message      string
	Iterations   int
	TestsPassed  bool
	TestCoverage float64
}

// workContext holds state shared between workflow steps.
type workContext struct {
	task         task.Task
	worktree     *worktree.Worktree
	branchName   string
	pinner       *contextpin.ContextPinner
	plan         *planner.Plan
	brainHandoff *brain.BrainHandoff
	collector    *brain.Collector
	exec         *executor.Executor
	execResult   *executor.ExecutionResult
	testResult   *testrunner.TestResult
	reviewResult *scottbott.ReviewResult
	iterations   int
	startTime    time.Time
	costTracker  *cost.Tracker
	draftPRURL   string // URL of draft PR created as safety checkpoint
}

// New creates a new Agent.
func New(cfg *config.Config) (*Agent, error) {
	return &Agent{
		config:       cfg,
		linearClient: linear.New(cfg.LinearKey),
		coordinator:  coordinator.New(),
	}, nil
}

// Work executes the complete workflow for a task.
// Orchestrates 9 steps: prepare → worktree → plan → validate → execute → test → review → commit → PR
func (a *Agent) Work(ctx context.Context, t task.Task) (*WorkResult, error) {
	wc := &workContext{
		task:        t,
		startTime:   time.Now(),
		costTracker: cost.NewTracker(),
	}

	// Start the coordinator
	a.coordinator.Start(ctx)
	defer a.coordinator.Stop()

	// Step 1: Prepare task (already received as parameter)
	if err := a.stepPrepareTask(ctx, wc); err != nil {
		return nil, err
	}

	// Step 2: Setup worktree
	if err := a.stepSetupWorktree(ctx, wc); err != nil {
		return nil, err
	}

	// Initialize brain collector for signal detection
	if a.config.Brain.Enabled {
		collector, err := brain.NewCollector(wc.worktree.Path)
		if err == nil {
			wc.collector = collector
			defer collector.Flush()
		}
	}

	// Step 3: Planning (also loads matching brains)
	if err := a.stepPlanning(ctx, wc); err != nil {
		return nil, err
	}

	// Step 4: Pre-flight validation
	if err := a.stepPreflightValidation(ctx, wc); err != nil {
		return nil, err
	}

	// Step 5: Execute development task
	if err := a.stepExecute(ctx, wc); err != nil {
		return nil, err
	}

	// Step 5b: Safety checkpoint — commit, push, and create a draft PR so work
	// is preserved even if test/review/refactor hangs or fails.
	if err := a.stepDraftPR(ctx, wc); err != nil {
		// Draft PR failure is non-fatal — log and continue
		fmt.Printf("   ⚠️  Draft PR checkpoint failed: %v\n", err)
	}

	// Step 6: Run tests and initial review (parallel)
	if err := a.stepTestAndReview(ctx, wc); err != nil {
		return nil, err
	}

	// Step 7: Review & refactor loop
	if err := a.stepRefactorLoop(ctx, wc); err != nil {
		return nil, err
	}

	// Post-workflow: auto-distill brains from accumulated signals
	if a.config.Brain.Enabled && brain.ShouldRunPeriodicDistill(24*time.Hour) {
		distiller := brain.NewAutoDistiller(wc.worktree.Path, a.config)
		results, err := distiller.DistillAll(ctx)
		if err != nil {
			fmt.Printf("   ⚠️  Auto-distillation error: %v\n", err)
		} else if len(results) > 0 {
			brain.RecordDistilled()
			for _, r := range results {
				method := "template"
				if r.UsedLLM {
					method = "LLM"
				}
				fmt.Printf("   🧠 Auto-generated brain: %s (%s, %d signals)\n", r.BrainID, method, r.Signals)
			}
		}
	}

	// Release context pins
	wc.pinner.Unpin("executor")

	// Check if review passed
	if !wc.reviewResult.Passed {
		result := &WorkResult{
			PRCreated:  false,
			Message:    "Review did not pass after max iterations",
			Iterations: wc.iterations,
		}
		if wc.draftPRURL != "" {
			result.PRURL = wc.draftPRURL
			result.PRCreated = true
			result.Message = "Review did not pass — draft PR preserved: " + wc.draftPRURL
		}
		return result, nil
	}

	// Step 8: Commit and push final reviewed changes
	if err := a.stepCommitAndPush(ctx, wc); err != nil {
		return nil, err
	}

	// Step 9: Finalize PR (update body with review info, mark ready)
	return a.stepFinalizePR(ctx, wc)
}

// ResumeWork resumes a previously failed execution from the review/refactor stage.
// It reuses the existing worktree (which contains the code changes from the failed run)
// and jumps directly to the test-and-review → refactor loop → commit → PR pipeline.
func (a *Agent) ResumeWork(ctx context.Context, t task.Task) (*WorkResult, error) {
	wc := &workContext{
		task:        t,
		startTime:   time.Now(),
		costTracker: cost.NewTracker(),
	}

	a.coordinator.Start(ctx)
	defer a.coordinator.Stop()

	// Step 1: Display task info
	if err := a.stepPrepareTask(ctx, wc); err != nil {
		return nil, err
	}

	// Step 2: Find and reuse existing worktree (instead of creating a new one)
	if err := a.stepResumeWorktree(ctx, wc); err != nil {
		return nil, err
	}

	// Initialize brain collector
	if a.config.Brain.Enabled {
		collector, err := brain.NewCollector(wc.worktree.Path)
		if err == nil {
			wc.collector = collector
			defer collector.Flush()
		}
	}

	// Skip planning — the code is already written.
	// Create executor to detect changed files.
	wc.exec = executor.New(wc.worktree.Path, a.config)
	changedFiles, err := wc.exec.DetectChangedFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to detect changed files in worktree: %w", err)
	}
	if len(changedFiles) == 0 {
		return nil, fmt.Errorf("no changed files found in worktree — nothing to resume")
	}

	fmt.Printf("   📁 Found %d changed files in worktree\n", len(changedFiles))
	for _, f := range changedFiles {
		fmt.Printf("      • %s\n", f)
	}
	fmt.Println()

	// Build a synthetic execution result from the existing worktree state.
	wc.execResult = &executor.ExecutionResult{
		Success:      true,
		FilesChanged: changedFiles,
	}

	// Stage changes
	fmt.Println("   📥 Staging changes...")
	if err := wc.exec.StageChanges(); err != nil {
		return nil, fmt.Errorf("failed to stage changes: %w", err)
	}

	// Safety checkpoint — ensure draft PR exists for resumed runs too
	if err := a.stepDraftPR(ctx, wc); err != nil {
		fmt.Printf("   ⚠️  Draft PR checkpoint failed: %v\n", err)
	}

	// Step 6: Run tests and review
	if err := a.stepTestAndReview(ctx, wc); err != nil {
		return nil, err
	}

	// Step 7: Review & refactor loop
	if err := a.stepRefactorLoop(ctx, wc); err != nil {
		return nil, err
	}

	// Release context pins
	wc.pinner.Unpin("executor")

	// Check if review passed
	if !wc.reviewResult.Passed {
		result := &WorkResult{
			PRCreated:  false,
			Message:    "Review did not pass after max iterations",
			Iterations: wc.iterations,
		}
		if wc.draftPRURL != "" {
			result.PRURL = wc.draftPRURL
			result.PRCreated = true
			result.Message = "Review did not pass — draft PR preserved: " + wc.draftPRURL
		}
		return result, nil
	}

	// Step 8: Commit and push
	if err := a.stepCommitAndPush(ctx, wc); err != nil {
		return nil, err
	}

	// Step 9: Finalize PR
	return a.stepFinalizePR(ctx, wc)
}

// stepResumeWorktree finds an existing worktree for the task's branch.
func (a *Agent) stepResumeWorktree(ctx context.Context, wc *workContext) error {
	agentID := fmt.Sprintf("resume-worktree-%s", wc.task.GetID())
	events.AgentStarted(agentID, "Resume Worktree", "Finding existing worktree")

	printStep(2, 9, "Resuming existing worktree")

	repoPath, err := os.Getwd()
	if err != nil {
		events.AgentCompleted(agentID, "Resume Worktree", "failed")
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	wtManager, err := worktree.New(repoPath)
	if err != nil {
		events.AgentCompleted(agentID, "Resume Worktree", "failed")
		return fmt.Errorf("failed to create worktree manager: %w", err)
	}

	branchName := wc.task.GetBranchName()
	fmt.Printf("   🌿 Looking for worktree with branch: %s\n", branchName)

	// List existing worktrees and find one matching this branch
	worktrees, err := wtManager.List()
	if err != nil {
		events.AgentCompleted(agentID, "Resume Worktree", "failed")
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	var found *worktree.Worktree
	for _, wt := range worktrees {
		if wt.BranchName == branchName {
			found = wt
			break
		}
	}

	if found == nil {
		events.AgentCompleted(agentID, "Resume Worktree", "failed")
		return fmt.Errorf("no existing worktree found for branch %s — cannot resume", branchName)
	}

	fmt.Printf("   ♻️  Found existing worktree: %s\n", found.Path)
	fmt.Println()

	wc.worktree = found
	wc.branchName = branchName
	wc.pinner = contextpin.New(found.Path)
	wc.pinner.SetCoordinator(a.coordinator)

	events.AgentCompletedWithData(agentID, "Resume Worktree", "success", map[string]any{
		"worktree_path": found.Path,
		"branch":        branchName,
		"resumed":       true,
	})
	return nil
}

// stepPrepareTask displays task information (Step 1).
func (a *Agent) stepPrepareTask(ctx context.Context, wc *workContext) error {
	agentID := fmt.Sprintf("prepare-%s", wc.task.GetID())
	events.AgentStarted(agentID, "Preparing Task", fmt.Sprintf("Preparing task %s", wc.task.GetID()))

	printStep(1, 9, "Preparing task")
	fmt.Printf("   🎫 Task ID: %s\n", wc.task.GetID())

	metadata := wc.task.GetMetadata()
	fmt.Printf("   📋 Source: %s\n", metadata.Source)
	fmt.Printf("   📝 Title: %s\n", wc.task.GetTitle())

	labels := wc.task.GetLabels()
	if len(labels) > 0 {
		fmt.Printf("   🏷️  Labels: %s\n", strings.Join(labels, ", "))
	}

	fmt.Println()
	fmt.Println("   📝 Description:")
	printIndented(truncate(wc.task.GetDescription(), 800), "      ")
	fmt.Println()

	events.AgentCompleted(agentID, "Preparing Task", "success")
	return nil
}

// stepSetupWorktree creates a git worktree for the task (Step 2).
func (a *Agent) stepSetupWorktree(ctx context.Context, wc *workContext) error {
	agentID := fmt.Sprintf("worktree-%s", wc.task.GetID())
	events.AgentStarted(agentID, "Setup Worktree", "Creating isolated git worktree")

	printStep(2, 9, "Setting up git worktree")

	repoPath, err := os.Getwd()
	if err != nil {
		events.AgentCompleted(agentID, "Setup Worktree", "failed")
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	fmt.Printf("   📂 Repo: %s\n", repoPath)

	wtManager, err := worktree.New(repoPath)
	if err != nil {
		events.AgentCompleted(agentID, "Setup Worktree", "failed")
		return fmt.Errorf("failed to create worktree manager: %w", err)
	}

	branchName := wc.task.GetBranchName()
	fmt.Printf("   🌿 Branch: %s\n", branchName)

	wt, err := wtManager.Create(branchName, a.config.BaseBranch)
	if err != nil {
		events.AgentCompleted(agentID, "Setup Worktree", "failed")
		return fmt.Errorf("failed to create worktree: %w", err)
	}
	fmt.Printf("   📁 Worktree: %s\n", wt.Path)
	fmt.Println()

	wc.worktree = wt
	wc.branchName = branchName

	// Initialize context pinner for multi-file coordination
	wc.pinner = contextpin.New(wt.Path)
	wc.pinner.SetCoordinator(a.coordinator)

	events.AgentCompletedWithData(agentID, "Setup Worktree", "success", map[string]any{
		"worktree_path": wt.Path,
		"branch":        branchName,
		"base_branch":   a.config.BaseBranch,
	})
	return nil
}

// stepPlanning runs the planning agent to analyze the task (Step 3).
// If a PreloadedPlan is set on the agent, it is used directly and the
// Claude planning call is skipped — saving tokens and time.
func (a *Agent) stepPlanning(ctx context.Context, wc *workContext) error {
	agentID := fmt.Sprintf("planning-%s", wc.task.GetID())

	printStep(3, 9, "Planning & analysis (parallel)")

	// Use pre-generated triage plan if available.
	if a.PreloadedPlan != nil {
		events.AgentStarted(agentID, "Planning & Analysis", "Using pre-generated triage plan")
		wc.plan = a.PreloadedPlan
		fmt.Printf("   📋 Using pre-generated plan: %s\n", truncate(a.PreloadedPlan.Summary, 120))
		fmt.Printf("   📁 %d candidate files, %d warnings\n", len(a.PreloadedPlan.RelevantFiles), len(a.PreloadedPlan.Warnings))
		events.AgentCompletedWithData(agentID, "Planning & Analysis", "success", map[string]any{
			"plan":       a.PreloadedPlan.Summary,
			"preloaded":  true,
		})
	} else {
		events.AgentStarted(agentID, "Planning & Analysis", "Analyzing codebase and creating implementation plan")

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			planAgent := planner.New(wc.worktree.Path, a.config)
			plan, usage, err := planAgent.Analyze(ctx, wc.task)
			if err != nil {
				fmt.Printf("   ⚠️  Planning failed: %v (continuing without plan)\n", err)
				events.AgentCompleted(agentID, "Planning & Analysis", "failed")
				return
			}
			wc.plan = plan
			if usage != nil {
				wc.costTracker.Add("Planning", *usage)
			}
			planText := ""
			if plan != nil {
				planText = plan.Summary
			}
			events.AgentCompletedWithData(agentID, "Planning & Analysis", "success", map[string]any{
				"plan": planText,
			})
		}()
		wg.Wait()
	}

	// Load matching brains based on task content and plan
	if a.config.Brain.Enabled {
		keywords := brain.ExtractKeywords(wc.task.GetTitle() + " " + wc.task.GetDescription())
		var filePaths []string
		if wc.plan != nil {
			filePaths = wc.plan.RelevantFiles
		}

		injector := brain.NewInjector(wc.worktree.Path, a.config.Brain.MaxBrains)
		brainHandoff, err := injector.ForContext(keywords, filePaths, nil)
		if err != nil {
			fmt.Printf("   ⚠️  Brain loading failed: %v\n", err)
		} else if brainHandoff != nil {
			wc.brainHandoff = brainHandoff
			fmt.Printf("   🧠 Loaded brain context: %s\n", brainHandoff.Concise())
		}
	}

	fmt.Println()

	return nil
}

// stepPreflightValidation validates the plan before execution (Step 4).
func (a *Agent) stepPreflightValidation(ctx context.Context, wc *workContext) error {
	agentID := fmt.Sprintf("preflight-%s", wc.task.GetID())
	events.AgentStarted(agentID, "Pre-flight Validation", "Validating implementation plan")

	printStep(4, 9, "Pre-flight validation")

	if wc.plan == nil {
		fmt.Println("   ⏭️  Skipping (no plan)")
		fmt.Println()
		events.AgentCompleted(agentID, "Pre-flight Validation", "success")
		return nil
	}

	preflightAgent := preflight.New(wc.worktree.Path)
	preflightAgent.SetCoordinator(a.coordinator)
	validation, err := preflightAgent.Validate(ctx, wc.plan)
	if err != nil {
		fmt.Printf("   ⚠️  Validation error: %v\n", err)
		events.AgentCompleted(agentID, "Pre-flight Validation", "failed")
	} else {
		fmt.Printf("   %s\n", (&preflight.ValidationHandoff{Result: validation}).Concise())
		if !validation.Valid {
			fmt.Println("   ⚠️  Validation failed but continuing...")
			for _, e := range validation.Errors {
				fmt.Printf("      ❌ %s\n", e.Message)
			}
		}
		for _, w := range validation.Warnings {
			fmt.Printf("      ⚠️  %s\n", w.Message)
		}
		events.AgentCompleted(agentID, "Pre-flight Validation", "success")
	}

	// Pin files from the plan for context consistency
	if len(wc.plan.RelevantFiles) > 0 {
		fmt.Println("   📌 Pinning context for relevant files...")
		wc.pinner.AnalyzeFiles(wc.plan.RelevantFiles)
		if _, err := wc.pinner.Pin("executor", wc.plan.RelevantFiles, false); err != nil {
			fmt.Printf("   ⚠️  Could not pin files: %v\n", err)
		}
	}

	fmt.Println()
	return nil
}

// stepExecute runs the executor to implement the task (Step 5).
func (a *Agent) stepExecute(ctx context.Context, wc *workContext) error {
	agentID := fmt.Sprintf("execute-%s", wc.task.GetID())
	events.AgentStarted(agentID, "Execution", "Implementing code changes")

	printStep(5, 9, "Executing development task")

	wc.exec = executor.New(wc.worktree.Path, a.config)

	// Inject brain context if available
	if wc.brainHandoff != nil {
		wc.exec.SetBrainHandoff(wc.brainHandoff, a.config.Brain.TokenBudget)
	}

	result, usage, err := wc.exec.ExecuteWithPlan(ctx, wc.task, wc.plan)
	if err != nil {
		events.AgentCompleted(agentID, "Execution", "failed")
		return fmt.Errorf("execution failed: %w", err)
	}

	if usage != nil {
		wc.costTracker.Add("Execution", *usage)
	}

	if !result.Success {
		events.AgentCompleted(agentID, "Execution", "failed")
		return fmt.Errorf("execution failed: %v", result.Error)
	}

	wc.execResult = result
	fmt.Println()

	// Stage changes
	fmt.Println("   📥 Staging changes...")
	if err := wc.exec.StageChanges(); err != nil {
		events.AgentCompleted(agentID, "Execution", "failed")
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Get diff for metadata
	diff, _ := wc.exec.GetDiff()
	events.AgentCompletedWithData(agentID, "Execution", "success", map[string]any{
		"diff": diff,
	})
	return nil
}

// stepTestAndReview runs tests and initial review in parallel (Step 6).
func (a *Agent) stepTestAndReview(ctx context.Context, wc *workContext) error {
	testAgentID := fmt.Sprintf("test-%s", wc.task.GetID())
	reviewAgentID := fmt.Sprintf("review-1-%s", wc.task.GetID())

	printStep(6, 9, "Running tests & initial review (parallel)")

	// Get diff for review
	initialDiff, err := wc.exec.GetDiff()
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Run tests in parallel with a timeout to prevent hanging
	go func() {
		defer wg.Done()
		events.AgentStarted(testAgentID, "Running Tests", "Running unit tests for changed files")
		testCtx, testCancel := context.WithTimeout(ctx, 5*time.Minute)
		defer testCancel()
		testAgent := testrunner.New(wc.worktree.Path)
		testAgent.SetCoordinator(a.coordinator)
		var testErr error
		wc.testResult, testErr = testAgent.RunForFiles(testCtx, wc.execResult.FilesChanged)
		if testErr != nil {
			fmt.Printf("   ⚠️  Test runner error: %v\n", testErr)
		}
		if wc.testResult != nil && wc.testResult.Passed {
			events.AgentCompleted(testAgentID, "Running Tests", "success")
		} else {
			events.AgentCompleted(testAgentID, "Running Tests", "failed")
		}
	}()

	// Run initial review in parallel
	go func() {
		defer wg.Done()
		events.AgentStarted(reviewAgentID, "Code Review #1", "Reviewing code quality and best practices")
		reviewHandoff := handoff.NewReviewHandoff(wc.task, initialDiff, wc.execResult.FilesChanged)
		reviewer := scottbott.NewWithSkill(wc.worktree.Path, 1, a.config.ReviewSkill, a.config)
		reviewResult, usage, _ := reviewer.Review(ctx, reviewHandoff.Concise(), initialDiff)
		wc.reviewResult = reviewResult
		if usage != nil {
			wc.costTracker.Add("Review #1", *usage)
		}
		if reviewResult != nil && reviewResult.Passed {
			feedback := reviewResult.Summary
			if feedback == "" && len(reviewResult.Issues) > 0 {
				feedback = fmt.Sprintf("Found %d issues", len(reviewResult.Issues))
			}
			events.AgentCompletedWithData(reviewAgentID, "Code Review #1", "success", map[string]any{
				"feedback": feedback,
				"issues":   reviewResult.Issues,
			})
		} else {
			feedback := ""
			if reviewResult != nil {
				feedback = reviewResult.Summary
			}
			events.AgentCompletedWithData(reviewAgentID, "Code Review #1", "failed", map[string]any{
				"feedback": feedback,
			})
		}
	}()

	wg.Wait()

	// Display test results
	if wc.testResult != nil {
		fmt.Printf("   🧪 Tests: %s\n", (&testrunner.TestResultHandoff{Result: wc.testResult}).Concise())
	}

	// Display review results
	if wc.reviewResult != nil {
		fmt.Println(wc.reviewResult.FormatReview())
	}
	fmt.Println()

	return nil
}

// stepRefactorLoop runs the review/refactor loop until passing or max iterations (Step 7).
func (a *Agent) stepRefactorLoop(ctx context.Context, wc *workContext) error {
	printStep(7, 9, "Review & refactor loop")

	previousDiff, _ := wc.exec.GetDiff()

	for wc.iterations < a.config.MaxIterations {
		wc.iterations++
		fmt.Printf("\n   🔄 Iteration %d of %d\n", wc.iterations, a.config.MaxIterations)
		fmt.Println("   ─────────────────────────────")

		events.Progress(fmt.Sprintf("Review & refactor iteration %d of %d", wc.iterations, a.config.MaxIterations))

		// Use existing review for first iteration, get fresh review for subsequent
		if wc.iterations > 1 || wc.reviewResult == nil {
			if err := a.doReview(ctx, wc, &previousDiff); err != nil {
				return err
			}
		}

		if wc.reviewResult.Passed {
			fmt.Println("   ✅ Review passed!")

			// Run final tests to confirm
			if wc.testResult == nil || !wc.testResult.Passed {
				testAgent := testrunner.New(wc.worktree.Path)
				wc.testResult, _ = testAgent.RunForFiles(ctx, wc.execResult.FilesChanged)
				if wc.testResult != nil && !wc.testResult.Passed {
					fmt.Printf("   ⚠️  Tests failed: %s\n", (&testrunner.TestResultHandoff{Result: wc.testResult}).Concise())
					wc.reviewResult.Passed = false
					wc.reviewResult.Issues = append(wc.reviewResult.Issues, scottbott.Issue{
						Severity:    "major",
						Description: fmt.Sprintf("Tests failed: %d failures", wc.testResult.FailedTests),
					})
				}
			}

			if wc.reviewResult.Passed {
				break
			}
		}

		// Collect review failure signals
		if wc.collector != nil && wc.reviewResult != nil && !wc.reviewResult.Passed {
			wc.collector.OnReviewFailure(
				wc.reviewResult.GetIssueDescriptions(),
				wc.execResult.FilesChanged,
			)
		}

		if wc.iterations >= a.config.MaxIterations {
			fmt.Println("   ⚠️  Maximum iterations reached without passing review")
			break
		}

		// Refactor based on feedback
		if err := a.doRefactor(ctx, wc, previousDiff); err != nil {
			return err
		}

		// Collect refactor signals
		if wc.collector != nil {
			wc.collector.OnRefactorIteration(
				wc.iterations,
				wc.reviewResult.GetIssueDescriptions(),
				wc.execResult.FilesChanged,
			)
		}
	}

	return nil
}

// doReview gets a fresh diff and runs the review.
func (a *Agent) doReview(ctx context.Context, wc *workContext, previousDiff *string) error {
	diff, err := wc.exec.GetDiff()
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}
	fmt.Printf("   📏 Diff size: %d lines\n", strings.Count(diff, "\n"))

	reviewHandoff := handoff.NewReviewHandoff(wc.task, diff, wc.execResult.FilesChanged)
	reviewer := scottbott.NewWithSkill(wc.worktree.Path, wc.iterations, a.config.ReviewSkill, a.config)
	reviewResult, usage, err := reviewer.Review(ctx, reviewHandoff.ForTokenBudget(handoff.DefaultBudget.Context), diff)
	if err != nil {
		return fmt.Errorf("review failed: %w", err)
	}

	if usage != nil {
		wc.costTracker.Add(fmt.Sprintf("Review #%d", wc.iterations), *usage)
	}

	fmt.Println(reviewResult.FormatReview())
	wc.reviewResult = reviewResult
	*previousDiff = diff

	return nil
}

// doRefactor performs refactoring based on review feedback.
func (a *Agent) doRefactor(ctx context.Context, wc *workContext, previousDiff string) error {
	refactorAgentID := fmt.Sprintf("refactor-%d-%s", wc.iterations, wc.task.GetID())
	events.AgentStarted(refactorAgentID, fmt.Sprintf("Refactoring #%d", wc.iterations), "Applying code review feedback")

	fmt.Printf("   🔧 Refactoring (attempt %d)...\n", wc.iterations)

	refactorExec := executor.NewRefactorExecutor(wc.worktree.Path, wc.iterations, a.config)
	currentCode, _ := refactorExec.GetSpecificFiles(wc.execResult.FilesChanged)

	// Load project rules for proper refactoring
	projectRules := refactorExec.LoadProjectRules()

	refactorHandoff := handoff.NewRefactorHandoff(
		wc.task,
		wc.reviewResult.GetIssueDescriptions(),
		wc.reviewResult.Guidance,
		wc.execResult.FilesChanged,
		currentCode,
		projectRules,
	)

	refactorResult, usage, err := refactorExec.RefactorWithHandoff(ctx, refactorHandoff)
	if err != nil {
		events.AgentCompleted(refactorAgentID, fmt.Sprintf("Refactoring #%d", wc.iterations), "failed")
		return fmt.Errorf("refactor failed: %w", err)
	}

	if usage != nil {
		wc.costTracker.Add(fmt.Sprintf("Refactor #%d", wc.iterations), *usage)
	}

	// Verify the diff addresses the issues
	if len(wc.reviewResult.Issues) > 0 {
		newDiff, _ := wc.exec.GetDiff()
		verifier := diffverify.New(wc.worktree.Path)
		verifier.SetCoordinator(a.coordinator)
		verification, _ := verifier.Verify(ctx, wc.reviewResult.Issues, previousDiff, newDiff)
		if verification != nil {
			fmt.Printf("   🔍 Verification: %s\n", (&diffverify.VerificationHandoff{Result: verification}).Concise())
			if len(verification.UnaddressedIssues) > 0 {
				fmt.Printf("   ⚠️  %d issues may not be addressed\n", len(verification.UnaddressedIssues))
			}
		}
	}

	wc.execResult.FilesChanged = refactorResult.FilesChanged

	if !refactorResult.Success {
		events.AgentCompleted(refactorAgentID, fmt.Sprintf("Refactoring #%d", wc.iterations), "failed")
		return fmt.Errorf("refactor failed: %v", refactorResult.Error)
	}

	// Stage new changes
	fmt.Println("   📥 Staging refactored changes...")
	if err := wc.exec.StageChanges(); err != nil {
		events.AgentCompleted(refactorAgentID, fmt.Sprintf("Refactoring #%d", wc.iterations), "failed")
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Get refactored diff for metadata
	refactorDiff, _ := wc.exec.GetDiff()
	events.AgentCompletedWithData(refactorAgentID, fmt.Sprintf("Refactoring #%d", wc.iterations), "success", map[string]any{
		"refactor_diff": refactorDiff,
	})
	return nil
}

// stepDraftPR creates a safety-checkpoint draft PR right after execution,
// before test/review/refactor. This preserves work even if later stages hang or fail.
func (a *Agent) stepDraftPR(ctx context.Context, wc *workContext) error {
	agentID := fmt.Sprintf("draft-pr-%s", wc.task.GetID())
	events.AgentStarted(agentID, "Draft PR", "Creating safety checkpoint draft PR")

	fmt.Println("   📋 Creating draft PR checkpoint...")

	// Stage and commit current state
	if err := wc.exec.StageChanges(); err != nil {
		events.AgentCompleted(agentID, "Draft PR", "failed")
		return fmt.Errorf("failed to stage changes for draft PR: %w", err)
	}

	commitMsg := fmt.Sprintf("wip(%s): %s [draft checkpoint]",
		wc.task.GetID(),
		wc.task.GetTitle(),
	)
	if err := wc.exec.Commit(commitMsg); err != nil {
		events.AgentCompleted(agentID, "Draft PR", "failed")
		return fmt.Errorf("failed to commit for draft PR: %w", err)
	}

	// Push
	if err := wc.exec.Push(wc.branchName); err != nil {
		events.AgentCompleted(agentID, "Draft PR", "failed")
		return fmt.Errorf("failed to push for draft PR: %w", err)
	}

	// Build draft body
	metadata := wc.task.GetMetadata()
	var draftBody string
	if metadata.Source == task.SourceLinear {
		draftBody = fmt.Sprintf("## %s\n\n### Ticket\n[%s](https://linear.app/issue/%s)\n\n"+
			"### Status\nDraft checkpoint — review/refactor in progress.\n\n---\n*Automated by BoatmanMode*",
			wc.task.GetTitle(), wc.task.GetID(), wc.task.GetID())
	} else {
		draftBody = fmt.Sprintf("## %s\n\n### Status\nDraft checkpoint — review/refactor in progress.\n\n---\n*Automated by BoatmanMode*",
			wc.task.GetTitle())
	}

	prResult, err := github.CreateDraftPRInDir(ctx, wc.worktree.Path, wc.task.GetTitle(), draftBody, a.config.BaseBranch)
	if err != nil {
		events.AgentCompleted(agentID, "Draft PR", "failed")
		return fmt.Errorf("failed to create draft PR: %w", err)
	}

	wc.draftPRURL = prResult.URL
	fmt.Printf("   📋 Draft PR: %s\n\n", prResult.URL)
	events.AgentCompleted(agentID, "Draft PR", "success")
	return nil
}

// stepFinalizePR updates the draft PR body with review results and marks it ready.
// If no draft PR exists, falls back to creating a new PR.
func (a *Agent) stepFinalizePR(ctx context.Context, wc *workContext) (*WorkResult, error) {
	if wc.draftPRURL == "" {
		// No draft PR — create from scratch (fallback)
		return a.stepCreatePR(ctx, wc)
	}

	agentID := fmt.Sprintf("finalize-pr-%s", wc.task.GetID())
	events.AgentStarted(agentID, "Finalize PR", "Updating draft PR and marking ready")

	printStep(9, 9, "Finalizing pull request")

	// Build final PR body with review info
	prBody := a.buildPRBody(wc)

	// Update PR body
	if err := github.UpdatePRBody(ctx, wc.worktree.Path, prBody); err != nil {
		fmt.Printf("   ⚠️  Failed to update PR body: %v\n", err)
	}

	// Mark PR ready
	fmt.Println("   ✅ Marking PR as ready for review...")
	if err := github.MarkPRReady(ctx, wc.worktree.Path); err != nil {
		events.AgentCompleted(agentID, "Finalize PR", "failed")
		return nil, fmt.Errorf("failed to mark PR ready: %w", err)
	}

	events.AgentCompleted(agentID, "Finalize PR", "success")
	a.printWorkflowSummary(wc, wc.draftPRURL)

	return &WorkResult{
		PRCreated:    true,
		PRURL:        wc.draftPRURL,
		Message:      "Successfully finalized PR",
		Iterations:   wc.iterations,
		TestsPassed:  wc.testResult == nil || wc.testResult.Passed,
		TestCoverage: getTestCoverage(wc.testResult),
	}, nil
}

// buildPRBody generates the full PR description with review results.
func (a *Agent) buildPRBody(wc *workContext) string {
	metadata := wc.task.GetMetadata()

	if metadata.Source == task.SourceLinear {
		return fmt.Sprintf(`## %s

### Ticket
[%s](https://linear.app/issue/%s)

### Description
%s

### Changes
%s

### Quality
- Review iterations: %d
- Tests: %s
- Coverage: %.1f%%

---
*Automated by BoatmanMode*
`,
			wc.task.GetTitle(),
			wc.task.GetID(), wc.task.GetID(),
			wc.task.GetDescription(),
			wc.reviewResult.Summary,
			wc.iterations,
			formatTestStatus(wc.testResult),
			getTestCoverage(wc.testResult),
		)
	}

	taskType := "Prompt-based"
	if metadata.Source == task.SourceFile {
		taskType = "File-based"
	}
	return fmt.Sprintf(`## %s

### Task
%s task (%s)

### Description
%s

### Changes
%s

### Quality
- Review iterations: %d
- Tests: %s
- Coverage: %.1f%%

---
*Automated by BoatmanMode*
`,
		wc.task.GetTitle(),
		taskType, wc.task.GetID(),
		truncate(wc.task.GetDescription(), 500),
		wc.reviewResult.Summary,
		wc.iterations,
		formatTestStatus(wc.testResult),
		getTestCoverage(wc.testResult),
	)
}

// stepCommitAndPush commits and pushes changes (Step 8).
func (a *Agent) stepCommitAndPush(ctx context.Context, wc *workContext) error {
	agentID := fmt.Sprintf("commit-%s", wc.task.GetID())
	events.AgentStarted(agentID, "Commit & Push", "Committing and pushing changes to remote")

	printStep(8, 9, "Committing and pushing")

	commitMsg := fmt.Sprintf("feat(%s): %s\n\n%s",
		wc.task.GetID(),
		wc.task.GetTitle(),
		wc.reviewResult.Summary,
	)
	fmt.Println("   💾 Staging and creating commit...")
	fmt.Printf("   📝 Message: %s\n", strings.Split(commitMsg, "\n")[0])

	if err := wc.exec.StageChanges(); err != nil {
		events.AgentCompleted(agentID, "Commit & Push", "failed")
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	if err := wc.exec.Commit(commitMsg); err != nil {
		events.AgentCompleted(agentID, "Commit & Push", "failed")
		return fmt.Errorf("failed to commit: %w", err)
	}

	fmt.Println("   📤 Pushing to origin...")
	if err := wc.exec.Push(wc.branchName); err != nil {
		events.AgentCompleted(agentID, "Commit & Push", "failed")
		return fmt.Errorf("failed to push: %w", err)
	}
	fmt.Println()

	events.AgentCompleted(agentID, "Commit & Push", "success")
	return nil
}

// stepCreatePR creates a pull request (Step 9).
func (a *Agent) stepCreatePR(ctx context.Context, wc *workContext) (*WorkResult, error) {
	agentID := fmt.Sprintf("pr-%s", wc.task.GetID())
	events.AgentStarted(agentID, "Create PR", "Creating pull request")

	printStep(9, 9, "Creating pull request")

	// Format PR body based on task source
	metadata := wc.task.GetMetadata()
	var prBody string

	if metadata.Source == task.SourceLinear {
		// Linear mode - include ticket link
		prBody = fmt.Sprintf(`## %s

### Ticket
[%s](https://linear.app/issue/%s)

### Description
%s

### Changes
%s

### Quality
- Review iterations: %d
- Tests: %s
- Coverage: %.1f%%

---
*Automated by BoatmanMode 🚣*
`,
			wc.task.GetTitle(),
			wc.task.GetID(),
			wc.task.GetID(),
			wc.task.GetDescription(),
			wc.reviewResult.Summary,
			wc.iterations,
			formatTestStatus(wc.testResult),
			getTestCoverage(wc.testResult),
		)
	} else {
		// Prompt/File mode - no ticket link
		taskType := "Prompt-based"
		if metadata.Source == task.SourceFile {
			taskType = "File-based"
		}

		prBody = fmt.Sprintf(`## %s

### Task
%s task (%s)

### Description
%s

### Changes
%s

### Quality
- Review iterations: %d
- Tests: %s
- Coverage: %.1f%%

---
*Automated by BoatmanMode 🚣*
`,
			wc.task.GetTitle(),
			taskType,
			wc.task.GetID(),
			truncate(wc.task.GetDescription(), 500),
			wc.reviewResult.Summary,
			wc.iterations,
			formatTestStatus(wc.testResult),
			getTestCoverage(wc.testResult),
		)
	}

	fmt.Println("   🔗 Running: gh pr create")
	prResult, err := github.CreatePRInDir(ctx, wc.worktree.Path, wc.task.GetTitle(), prBody, a.config.BaseBranch)
	if err != nil {
		events.AgentCompleted(agentID, "Create PR", "failed")
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	events.AgentCompleted(agentID, "Create PR", "success")
	a.printWorkflowSummary(wc, prResult.URL)

	return &WorkResult{
		PRCreated:    true,
		PRURL:        prResult.URL,
		Message:      "Successfully created PR",
		Iterations:   wc.iterations,
		TestsPassed:  wc.testResult == nil || wc.testResult.Passed,
		TestCoverage: getTestCoverage(wc.testResult),
	}, nil
}

// printWorkflowSummary prints the final workflow completion summary.
func (a *Agent) printWorkflowSummary(wc *workContext, prURL string) {
	totalElapsed := time.Since(wc.startTime)

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════════════")
	fmt.Println("✅ WORKFLOW COMPLETE")
	fmt.Println("═══════════════════════════════════════════════════════════════════════")
	fmt.Printf("   🎫 Task:       %s\n", wc.task.GetID())
	fmt.Printf("   🌿 Branch:     %s\n", wc.branchName)
	fmt.Printf("   🔄 Iterations: %d\n", wc.iterations)
	fmt.Printf("   🧪 Tests:      %s\n", formatTestStatus(wc.testResult))
	fmt.Printf("   ⏱️  Total time: %s\n", totalElapsed.Round(time.Second))
	fmt.Printf("   🔗 PR:         %s\n", prURL)

	// Display cost summary if any usage was tracked
	if wc.costTracker.HasUsage() {
		fmt.Print(wc.costTracker.Summary())
	}

	fmt.Println("═══════════════════════════════════════════════════════════════════════")
}

// formatTestStatus formats test result for display.
func formatTestStatus(result *testrunner.TestResult) string {
	if result == nil {
		return "N/A"
	}
	if result.Passed {
		return fmt.Sprintf("✅ %d passed", result.PassedTests)
	}
	return fmt.Sprintf("❌ %d failed, %d passed", result.FailedTests, result.PassedTests)
}

// getTestCoverage extracts coverage from test result.
func getTestCoverage(result *testrunner.TestResult) float64 {
	if result == nil {
		return 0
	}
	return result.Coverage
}

// printStep prints a formatted step header.
func printStep(current, total int, description string) {
	fmt.Println()
	fmt.Printf("━━━ Step %d/%d: %s ━━━\n", current, total, description)
}

// printIndented prints text with indentation.
func printIndented(text, indent string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		fmt.Printf("%s%s\n", indent, line)
	}
}

// getRepoURL gets the remote URL for the repository.
func getRepoURL(repoPath string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// truncate shortens a string to the given length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// sanitize makes a string safe for use in branch names.
func sanitize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, ":", "")
	if len(s) > 30 {
		s = s[:30]
	}
	return s
}

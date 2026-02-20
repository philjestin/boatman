// Package agent provides Claude-powered enhancement of scaffold stubs.
// It shells out to the claude CLI to replace placeholder implementations
// with real LLM-calling code.
package agent

import (
	"fmt"

	"github.com/philjestin/boatman-ecosystem/harness/scaffold"
)

// rolePrompt defines the prompt template for a specific role file.
type rolePrompt struct {
	// filename is the stub file to enhance (e.g. "developer.go").
	filename string

	// buildPrompt returns the full prompt given the config and current stub code.
	buildPrompt func(cfg *EnhanceConfig, stubCode string) string
}

// rolePrompts returns the prompts for all role files that should be enhanced.
func rolePrompts(cfg *EnhanceConfig) []rolePrompt {
	return []rolePrompt{
		{
			filename: "developer.go",
			buildPrompt: func(cfg *EnhanceConfig, stub string) string {
				return fmt.Sprintf(`You are enhancing a Go file that implements the runner.Developer interface for an AI agent harness.

The Developer interface has two methods:
- Execute(ctx context.Context, req *runner.Request, plan *runner.Plan) (*runner.ExecuteResult, error)
- Refactor(ctx context.Context, req *runner.Request, issues []review.Issue, guidance string, prevResult *runner.ExecuteResult) (*runner.RefactorResult, error)

The target LLM provider is: %s
The target project language is: %s

Replace the TODO stubs with a working implementation that:
1. Calls the %s API/CLI to generate code changes based on the request
2. Applies the changes to files in d.WorkDir
3. Returns the list of changed files, a unified diff, and a summary

For the Refactor method, include the review issues and guidance in the prompt to the LLM.

Keep the implementation simple and focused. Use the standard library and exec.Command where possible.
Do not add external dependencies beyond what's in go.mod.

Here is the current stub code:

%s

Return ONLY the complete Go file content, no markdown fences or explanation.`,
					cfg.Provider, cfg.ProjectLang, cfg.Provider, stub)
			},
		},
		{
			filename: "reviewer.go",
			buildPrompt: func(cfg *EnhanceConfig, stub string) string {
				return fmt.Sprintf(`You are enhancing a Go file that implements the review.Reviewer interface for an AI agent harness.

The Reviewer interface has one method:
- Review(ctx context.Context, diff string, context string) (*review.ReviewResult, error)

The target LLM provider is: %s

Replace the TODO stub with a working implementation that:
1. Sends the diff to the %s API/CLI for code review
2. Parses the response into a review.ReviewResult with:
   - Passed: whether the code passes review
   - Score: quality score 0-10
   - Summary: brief review summary
   - Issues: specific issues found (severity, file, description, suggestion)

Keep the implementation simple. Use the standard library and exec.Command where possible.

Here is the current stub code:

%s

Return ONLY the complete Go file content, no markdown fences or explanation.`,
					cfg.Provider, cfg.Provider, stub)
			},
		},
		{
			filename: "planner.go",
			buildPrompt: func(cfg *EnhanceConfig, stub string) string {
				return fmt.Sprintf(`You are enhancing a Go file that implements the runner.Planner interface for an AI agent harness.

The Planner interface has one method:
- Plan(ctx context.Context, req *runner.Request) (*runner.Plan, error)

The target LLM provider is: %s

Replace the TODO stub with a working implementation that:
1. Sends the request to the %s API/CLI to create an implementation plan
2. Parses the response into a runner.Plan with summary, steps, and relevant files

Keep the implementation simple. Use the standard library and exec.Command where possible.

Here is the current stub code:

%s

Return ONLY the complete Go file content, no markdown fences or explanation.`,
					cfg.Provider, cfg.Provider, stub)
			},
		},
		{
			filename: "tester.go",
			buildPrompt: func(cfg *EnhanceConfig, stub string) string {
				testCmd := scaffold.LangTestCommand(cfg.ProjectLang)
				return fmt.Sprintf(`You are enhancing a Go file that implements the runner.Tester interface for an AI agent harness.

The Tester interface has one method:
- Test(ctx context.Context, req *runner.Request, changedFiles []string) (*runner.TestResult, error)

The target project language is: %s
The suggested test command is: %s

Replace the TODO stub with a working implementation that:
1. Runs the test command (%s) using exec.Command
2. Parses the output to determine pass/fail
3. Extracts failed test names if any
4. Reports coverage if available

Keep the implementation simple. Use the standard library and exec.Command.

Here is the current stub code:

%s

Return ONLY the complete Go file content, no markdown fences or explanation.`,
					cfg.ProjectLang, testCmd, testCmd, stub)
			},
		},
	}
}

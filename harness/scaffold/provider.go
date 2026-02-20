package scaffold

// ProviderMeta holds provider-specific metadata used by templates.
type ProviderMeta struct {
	// Name is a human-readable name for the provider.
	Name string

	// ExecuteHint is a comment inserted in the Developer.Execute stub.
	ExecuteHint string

	// RefactorHint is a comment inserted in the Developer.Refactor stub.
	RefactorHint string

	// ReviewHint is a comment inserted in the Reviewer.Review stub.
	ReviewHint string

	// PlanHint is a comment inserted in the Planner.Plan stub.
	PlanHint string

	// TestHint is a comment inserted in the Tester.Test stub.
	TestHint string
}

// providerMeta maps each provider to its metadata.
var providerMeta = map[LLMProvider]ProviderMeta{
	ProviderClaude: {
		Name: "Claude (Anthropic)",
		ExecuteHint: `// Call the Anthropic API or shell out to the claude CLI:
	//   claude -p --output-format json "Implement: <req.Description>"
	// Parse the response and apply changes to files in d.WorkDir.`,
		RefactorHint: `// Call Claude with the review issues and guidance to fix the code:
	//   claude -p "Fix these issues: <issues>" --output-format json
	// Apply the changes and return the updated diff.`,
		ReviewHint: `// Use Claude to review the diff:
	//   claude -p "Review this diff for quality: <diff>"
	// Parse the response into review.ReviewResult.`,
		PlanHint: `// Ask Claude to create an implementation plan:
	//   claude -p "Plan: <req.Description>"
	// Parse the response into runner.Plan with steps and relevant files.`,
		TestHint: `// Run tests using the appropriate test command for your project.
	// For Go: exec.Command("go", "test", "./...")
	// Parse output to determine pass/fail and coverage.`,
	},
	ProviderOpenAI: {
		Name: "OpenAI",
		ExecuteHint: `// Call the OpenAI Chat Completions API:
	//   POST https://api.openai.com/v1/chat/completions
	//   Model: "gpt-4o" (or your preferred model)
	// Parse the response and apply changes to files in d.WorkDir.`,
		RefactorHint: `// Call OpenAI with the review issues to fix the code.
	// Include the issues and guidance in the system/user messages.
	// Apply the changes and return the updated diff.`,
		ReviewHint: `// Use OpenAI to review the diff:
	//   Send the diff as a user message with review instructions.
	// Parse the response into review.ReviewResult.`,
		PlanHint: `// Ask OpenAI to create an implementation plan.
	// Send the request description and parse the structured response.`,
		TestHint: `// Run tests using the appropriate test command for your project.
	// Parse output to determine pass/fail and coverage.`,
	},
	ProviderOllama: {
		Name: "Ollama (local)",
		ExecuteHint: `// Call your local Ollama instance:
	//   POST http://localhost:11434/api/generate
	//   Model: "codellama" or your preferred model
	// Parse the response and apply changes to files in d.WorkDir.`,
		RefactorHint: `// Call Ollama with the review issues to fix the code.
	// Include issues and guidance in the prompt.
	// Apply the changes and return the updated diff.`,
		ReviewHint: `// Use Ollama to review the diff.
	// Send the diff with review instructions to your local model.
	// Parse the response into review.ReviewResult.`,
		PlanHint: `// Ask Ollama to create an implementation plan.
	// Send the request description to your local model.`,
		TestHint: `// Run tests using the appropriate test command for your project.
	// Parse output to determine pass/fail and coverage.`,
	},
	ProviderGeneric: {
		Name: "Generic LLM",
		ExecuteHint: `// Call your LLM of choice to implement the requested changes.
	// Parse the response and apply changes to files in d.WorkDir.`,
		RefactorHint: `// Call your LLM with the review issues and guidance.
	// Apply the changes and return the updated diff.`,
		ReviewHint: `// Use your LLM to review the diff.
	// Parse the response into review.ReviewResult.`,
		PlanHint: `// Ask your LLM to create an implementation plan.
	// Parse the response into runner.Plan.`,
		TestHint: `// Run tests using the appropriate test command for your project.
	// Parse output to determine pass/fail and coverage.`,
	},
}

// getProviderMeta returns metadata for a provider, falling back to generic.
func getProviderMeta(p LLMProvider) ProviderMeta {
	if m, ok := providerMeta[p]; ok {
		return m
	}
	return providerMeta[ProviderGeneric]
}

// LangTestCommand returns a suggested test command for the project language.
func LangTestCommand(lang ProjectLanguage) string {
	switch lang {
	case LangGo:
		return "go test ./..."
	case LangTypeScript:
		return "npm test"
	case LangPython:
		return "pytest"
	case LangRuby:
		return "bundle exec rspec"
	default:
		return "make test"
	}
}

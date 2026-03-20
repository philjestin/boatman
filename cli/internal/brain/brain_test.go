package brain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	content := `---
domain: Employer Upsell
keywords:
  - upsell
  - premium
  - subscription
labels:
  - employer
paths:
  - app/models/employer
---

## Planning
Architecture context here.

## Execution
Implementation patterns here.
`

	meta, body, err := parseFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Domain != "Employer Upsell" {
		t.Errorf("expected domain 'Employer Upsell', got %q", meta.Domain)
	}
	if len(meta.Keywords) != 3 {
		t.Errorf("expected 3 keywords, got %d", len(meta.Keywords))
	}
	if len(meta.Labels) != 1 {
		t.Errorf("expected 1 label, got %d", len(meta.Labels))
	}
	if len(meta.Paths) != 1 {
		t.Errorf("expected 1 path, got %d", len(meta.Paths))
	}
	if body == "" {
		t.Error("expected non-empty body")
	}
}

func TestParseFrontmatterMissing(t *testing.T) {
	content := "# Just a heading\n\nSome content."
	meta, body, err := parseFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Domain != "" {
		t.Errorf("expected empty domain, got %q", meta.Domain)
	}
	if body != content {
		t.Error("expected body to equal full content when no frontmatter")
	}
}

func TestParseSections(t *testing.T) {
	body := `## Planning
Plan stuff here.
More planning.

## Execution
Execute stuff here.

## Review
Review stuff here.`

	sections := parseSections(body)

	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}

	if _, ok := sections["planning"]; !ok {
		t.Error("missing 'planning' section")
	}
	if _, ok := sections["execution"]; !ok {
		t.Error("missing 'execution' section")
	}
	if _, ok := sections["review"]; !ok {
		t.Error("missing 'review' section")
	}
}

func TestBrainForPhase(t *testing.T) {
	b := &Brain{
		Name: "test",
		Meta: BrainMeta{Domain: "Test Domain"},
		Sections: map[string]string{
			"planning":  "Plan content",
			"execution": "Exec content",
			"review":    "Review content",
		},
		Raw: "Full content",
	}

	tests := []struct {
		phase    string
		contains string
	}{
		{"planning", "Plan content"},
		{"execution", "Exec content"},
		{"review", "Review content"},
		{"refactor", "Review content"}, // alias fallback
		{"unknown", "Full content"},    // raw fallback
	}

	for _, tt := range tests {
		result := b.ForPhase(tt.phase)
		if result == "" {
			t.Errorf("ForPhase(%q) returned empty", tt.phase)
			continue
		}
		if tt.phase != "unknown" && !contains(result, tt.contains) {
			t.Errorf("ForPhase(%q) = %q, want to contain %q", tt.phase, result, tt.contains)
		}
	}
}

func TestBrainMatches(t *testing.T) {
	b := &Brain{
		Meta: BrainMeta{
			Keywords: []string{"upsell", "premium"},
			Labels:   []string{"employer"},
			Paths:    []string{"app/models/employer"},
		},
	}

	tests := []struct {
		name      string
		text      string
		labels    []string
		paths     []string
		expectHit bool
	}{
		{"keyword match", "implement upsell flow", nil, nil, true},
		{"label match", "something", []string{"employer"}, nil, true},
		{"path match", "something", nil, []string{"app/models/employer/subscription.rb"}, true},
		{"no match", "unrelated task", []string{"student"}, []string{"app/models/student/"}, false},
		{"case insensitive keyword", "Premium Feature", nil, nil, true},
		{"case insensitive label", "something", []string{"Employer"}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := b.matches(tt.text, tt.labels, tt.paths)
			if got != tt.expectHit {
				t.Errorf("matches() = %v, want %v", got, tt.expectHit)
			}
		})
	}
}

func TestLoaderMatch(t *testing.T) {
	// Create temp brains directory
	tmpDir := t.TempDir()
	brainsDir := filepath.Join(tmpDir, "brains")
	os.MkdirAll(brainsDir, 0755)

	// Write a brain file
	brainContent := `---
domain: Payments
keywords:
  - payment
  - billing
  - invoice
labels:
  - payments
paths:
  - app/models/payment
---

## Planning
Check the payment gateway integration.

## Execution
Always use the PaymentService, never hit the gateway directly.

## Review
Verify PCI compliance in all payment-related changes.
`
	os.WriteFile(filepath.Join(brainsDir, "payments.md"), []byte(brainContent), 0644)

	// Write an example file (should be skipped)
	os.WriteFile(filepath.Join(brainsDir, "_example.md"), []byte("# Example\nIgnore me."), 0644)

	loader := NewLoader(tmpDir)

	// Should match on keyword
	matched, err := loader.Match("Fix payment processing bug", "The billing system fails", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matched) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matched))
	}
	if matched[0].Meta.Domain != "Payments" {
		t.Errorf("expected domain 'Payments', got %q", matched[0].Meta.Domain)
	}

	// Should not match
	matched, err = loader.Match("Update user profile", "Change the avatar", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matched) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matched))
	}
}

func TestLoaderNoBrainsDir(t *testing.T) {
	loader := NewLoader(t.TempDir())
	brains, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(brains) != 0 {
		t.Errorf("expected 0 brains, got %d", len(brains))
	}
}

func TestMergeForPhase(t *testing.T) {
	brains := []*Brain{
		{
			Name: "a",
			Meta: BrainMeta{Domain: "Alpha"},
			Sections: map[string]string{
				"planning": "Alpha planning",
			},
			Raw: "Alpha full",
		},
		{
			Name: "b",
			Meta: BrainMeta{Domain: "Beta"},
			Sections: map[string]string{
				"planning": "Beta planning",
			},
			Raw: "Beta full",
		},
	}

	result := MergeForPhase(brains, "planning")
	if !contains(result, "Alpha planning") || !contains(result, "Beta planning") {
		t.Errorf("MergeForPhase did not include both brains: %q", result)
	}
	if !contains(result, "Domain Brain: Alpha") || !contains(result, "Domain Brain: Beta") {
		t.Errorf("MergeForPhase missing brain headers: %q", result)
	}
}

func TestMergeForPhaseEmpty(t *testing.T) {
	result := MergeForPhase(nil, "planning")
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

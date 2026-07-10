package verifier

import (
	"context"
	"testing"
)

func TestPolicyVerifierPassesCleanDiff(t *testing.T) {
	verdict, err := NewPolicyVerifier().Verify(context.Background(), Request{
		Diff:         "+fmt.Println(\"hello\")",
		ChangedFiles: []string{"internal/foo.go"},
	})
	if err != nil {
		t.Fatalf("Verify error: %v", err)
	}
	if verdict.Status != StatusPass || len(verdict.Findings) != 0 {
		t.Fatalf("verdict = %#v, want clean pass", verdict)
	}
}

func TestPolicyVerifierFailsSecretLikeDiff(t *testing.T) {
	verdict, err := NewPolicyVerifier().Verify(context.Background(), Request{
		Diff:         `+OPENAI_API_KEY="sk-test-secret-value"`,
		ChangedFiles: []string{".env"},
	})
	if err != nil {
		t.Fatalf("Verify error: %v", err)
	}
	if verdict.Status != StatusFail {
		t.Fatalf("Status = %q, want fail", verdict.Status)
	}
	if len(verdict.Findings) == 0 || verdict.Findings[0].Code != "secret-like-diff" {
		t.Fatalf("findings = %#v, want secret finding", verdict.Findings)
	}
}

func TestPolicyVerifierWarnsOnSensitivePathsWithoutFailing(t *testing.T) {
	verdict, err := NewPolicyVerifier().Verify(context.Background(), Request{
		Diff:         "+func Login() {}",
		ChangedFiles: []string{"internal/auth/session.go", "db/migrate/20260710_add_users.sql"},
	})
	if err != nil {
		t.Fatalf("Verify error: %v", err)
	}
	if verdict.Status != StatusPass {
		t.Fatalf("Status = %q, want pass with warnings", verdict.Status)
	}
	if len(verdict.Findings) != 2 {
		t.Fatalf("findings = %#v, want two warnings", verdict.Findings)
	}
}

func TestPolicyVerifierHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewPolicyVerifier().Verify(ctx, Request{}); err == nil {
		t.Fatal("Verify should return context error")
	}
}

package approvals

import (
	"context"
	"strings"
	"testing"
)

type classifierFunc func(context.Context, Action) (Decision, error)

func (f classifierFunc) Evaluate(ctx context.Context, action Action) (Decision, error) {
	return f(ctx, action)
}

func TestServiceAllowsLowRiskAction(t *testing.T) {
	decision, err := NewService().Evaluate(context.Background(), Action{
		ID:    "read-doc",
		Kind:  "read",
		Paths: []string{"docs/architecture.md"},
	})
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	if decision.Kind != DecisionAllow || decision.Risk != RiskLow {
		t.Fatalf("decision = %#v, want low-risk allow", decision)
	}
}

func TestServiceRequiresHumanForAuthPaymentAndMigrationChanges(t *testing.T) {
	action := Action{
		ID:    "change-auth",
		Kind:  "edit",
		Paths: []string{"internal/auth/session.go", "db/migrate/20260710_add_billing.sql"},
	}
	decision, err := NewService().Evaluate(context.Background(), action)
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	if decision.Kind != DecisionNeedsHuman {
		t.Fatalf("decision.Kind = %q, want needs_human", decision.Kind)
	}
	if !contains(decision.RuleIDs, "auth-or-payments") || !contains(decision.RuleIDs, "database-migration") {
		t.Fatalf("rules = %#v, want auth and migration gates", decision.RuleIDs)
	}
}

func TestServiceDeniesDestructiveCommand(t *testing.T) {
	decision, err := NewService().Evaluate(context.Background(), Action{
		ID:      "bad",
		Kind:    "bash",
		Command: "git reset --hard HEAD",
	})
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	if decision.Kind != DecisionDeny || decision.Risk != RiskCritical {
		t.Fatalf("decision = %#v, want critical deny", decision)
	}
}

func TestClassifierCanRaiseButNotLower(t *testing.T) {
	raise := classifierFunc(func(context.Context, Action) (Decision, error) {
		return Decision{Kind: DecisionNeedsHuman, Risk: RiskMedium, Reasons: []string{"uncertain blast radius"}}, nil
	})
	decision, err := NewService(WithClassifier(raise)).Evaluate(context.Background(), Action{ID: "gray", Kind: "edit", Paths: []string{"internal/foo.go"}})
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	if decision.Kind != DecisionNeedsHuman || !decision.ClassifierRaised {
		t.Fatalf("decision = %#v, want classifier-raised human gate", decision)
	}

	lower := classifierFunc(func(context.Context, Action) (Decision, error) {
		return Decision{Kind: DecisionAllow, Risk: RiskLow}, nil
	})
	decision, err = NewService(WithClassifier(lower)).Evaluate(context.Background(), Action{ID: "secret", Kind: "edit", Paths: []string{".env"}})
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	if decision.Kind != DecisionNeedsHuman || decision.ClassifierRaised {
		t.Fatalf("decision = %#v, deterministic gate should not be lowered", decision)
	}
}

func TestApprovalRequestLifecycle(t *testing.T) {
	store := NewInMemoryStore()
	request := NewRequest(Action{ID: "write-file", Kind: "edit"}, Decision{Kind: DecisionNeedsHuman, Risk: RiskMedium})
	created, err := store.Create(context.Background(), request)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if created.State != RequestPending || !strings.HasPrefix(created.ID, "approval-") {
		t.Fatalf("created = %#v, want pending approval request", created)
	}
	resolved, err := store.Resolve(context.Background(), created.ID, RequestApproved, "philip", "looks good")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}
	if resolved.State != RequestApproved || resolved.ResolvedAt == nil || resolved.ResolvedBy != "philip" {
		t.Fatalf("resolved = %#v, want approved request", resolved)
	}
	if _, err := store.Resolve(context.Background(), created.ID, RequestDenied, "philip", "late"); err == nil {
		t.Fatal("second resolve should fail")
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

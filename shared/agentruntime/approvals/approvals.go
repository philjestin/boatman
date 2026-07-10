// Package approvals implements Boatman's deterministic approval policy decision
// point for side-effecting agent actions.
package approvals

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DecisionKind is the normalized PDP outcome.
type DecisionKind string

const (
	DecisionAllow      DecisionKind = "allow"
	DecisionDeny       DecisionKind = "deny"
	DecisionNeedsHuman DecisionKind = "needs_human"
)

// Risk is a coarse risk level attached to approval decisions.
type Risk string

const (
	RiskLow      Risk = "low"
	RiskMedium   Risk = "medium"
	RiskHigh     Risk = "high"
	RiskCritical Risk = "critical"
)

// Action is the provider-neutral thing an agent wants to do.
type Action struct {
	ID       string            `json:"id,omitempty"`
	Name     string            `json:"name,omitempty"`
	Kind     string            `json:"kind,omitempty"`
	Actor    string            `json:"actor,omitempty"`
	Subject  string            `json:"subject,omitempty"`
	Command  string            `json:"command,omitempty"`
	Paths    []string          `json:"paths,omitempty"`
	Payload  string            `json:"payload,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Decision explains why an action is allowed, denied, or requires a person.
type Decision struct {
	ActionID         string       `json:"actionId,omitempty"`
	Kind             DecisionKind `json:"kind"`
	Risk             Risk         `json:"risk"`
	Reasons          []string     `json:"reasons,omitempty"`
	RuleIDs          []string     `json:"ruleIds,omitempty"`
	ClassifierRaised bool         `json:"classifierRaised,omitempty"`
}

// Rule is a deterministic policy rule.
type Rule struct {
	ID          string
	Description string
	Decision    DecisionKind
	Risk        Risk
	Reason      string
	Match       func(Action) bool
}

// Classifier can raise an otherwise allowed action to a human gate. It cannot
// lower deterministic rules or issue final denies.
type Classifier interface {
	Evaluate(ctx context.Context, action Action) (Decision, error)
}

// Service evaluates actions against deterministic rules plus an optional
// classifier used only for gray-zone escalation.
type Service struct {
	rules      []Rule
	classifier Classifier
}

// Option customizes an approval service.
type Option func(*Service)

// WithRules replaces the deterministic rule set.
func WithRules(rules []Rule) Option {
	return func(s *Service) {
		s.rules = append([]Rule(nil), rules...)
	}
}

// WithClassifier attaches a raise-only classifier.
func WithClassifier(classifier Classifier) Option {
	return func(s *Service) {
		s.classifier = classifier
	}
}

// NewService creates an approval PDP service.
func NewService(opts ...Option) *Service {
	service := &Service{rules: DefaultRules()}
	for _, opt := range opts {
		opt(service)
	}
	return service
}

// Evaluate returns the approval decision for an action.
func (s *Service) Evaluate(ctx context.Context, action Action) (Decision, error) {
	if err := ctx.Err(); err != nil {
		return Decision{}, err
	}
	decision := Decision{
		ActionID: action.ID,
		Kind:     DecisionAllow,
		Risk:     RiskLow,
	}
	for _, rule := range s.rules {
		if rule.Match == nil || !rule.Match(action) {
			continue
		}
		decision.RuleIDs = append(decision.RuleIDs, rule.ID)
		if rule.Reason != "" {
			decision.Reasons = append(decision.Reasons, rule.Reason)
		}
		decision.Risk = maxRisk(decision.Risk, rule.Risk)
		decision.Kind = maxDecision(decision.Kind, rule.Decision)
	}
	if decision.Kind != DecisionAllow || s.classifier == nil {
		return normalizeDecision(decision), nil
	}
	classified, err := s.classifier.Evaluate(ctx, action)
	if err != nil {
		return Decision{}, err
	}
	if classified.Kind == DecisionNeedsHuman {
		decision.Kind = DecisionNeedsHuman
		decision.Risk = maxRisk(decision.Risk, classified.Risk)
		decision.Reasons = append(decision.Reasons, classified.Reasons...)
		decision.RuleIDs = append(decision.RuleIDs, classified.RuleIDs...)
		decision.ClassifierRaised = true
	}
	return normalizeDecision(decision), nil
}

// DefaultRules returns the built-in deterministic approval policy.
func DefaultRules() []Rule {
	return []Rule{
		{
			ID:       "destructive-command",
			Decision: DecisionDeny,
			Risk:     RiskCritical,
			Reason:   "destructive command requires a hard stop",
			Match: func(action Action) bool {
				text := actionText(action)
				return strings.Contains(text, "rm -rf /") ||
					strings.Contains(text, "git reset --hard") ||
					strings.Contains(text, "drop database") ||
					strings.Contains(text, "drop table")
			},
		},
		{
			ID:       "secrets-or-credentials",
			Decision: DecisionNeedsHuman,
			Risk:     RiskHigh,
			Reason:   "secret or credential material requires human approval",
			Match: func(action Action) bool {
				text := actionText(action)
				for _, marker := range []string{".env", "secret", "credential", "private key", "api_key", "token=", "ghp_"} {
					if strings.Contains(text, marker) {
						return true
					}
				}
				return false
			},
		},
		{
			ID:       "auth-or-payments",
			Decision: DecisionNeedsHuman,
			Risk:     RiskHigh,
			Reason:   "auth, identity, payment, or billing changes require human approval",
			Match: func(action Action) bool {
				text := actionText(action)
				for _, marker := range []string{"auth", "oauth", "session", "password", "permission", "payment", "billing", "stripe"} {
					if strings.Contains(text, marker) {
						return true
					}
				}
				return false
			},
		},
		{
			ID:       "database-migration",
			Decision: DecisionNeedsHuman,
			Risk:     RiskMedium,
			Reason:   "database migrations require human approval",
			Match: func(action Action) bool {
				text := actionText(action)
				return strings.Contains(text, "migration") ||
					strings.Contains(text, "db/migrate") ||
					strings.Contains(text, "schema.sql")
			},
		},
		{
			ID:       "external-write",
			Decision: DecisionNeedsHuman,
			Risk:     RiskMedium,
			Reason:   "external service writes require human approval",
			Match: func(action Action) bool {
				kind := strings.ToLower(strings.TrimSpace(action.Kind))
				if kind == "read" || kind == "inspect" || kind == "search" {
					return false
				}
				text := strings.ToLower(action.Kind + " " + action.Subject + " " + action.Name)
				for _, marker := range []string{"github", "linear", "slack", "datadog", "pagerduty"} {
					if strings.Contains(text, marker) {
						return true
					}
				}
				return false
			},
		},
		{
			ID:       "large-change",
			Decision: DecisionNeedsHuman,
			Risk:     RiskMedium,
			Reason:   "large changes require human approval",
			Match: func(action Action) bool {
				if action.Metadata == nil {
					return false
				}
				files, err := strconv.Atoi(action.Metadata["filesChanged"])
				return err == nil && files >= 25
			},
		},
	}
}

// RequestState is the lifecycle state for a durable approval request.
type RequestState string

const (
	RequestPending  RequestState = "pending"
	RequestApproved RequestState = "approved"
	RequestDenied   RequestState = "denied"
	RequestCanceled RequestState = "canceled"
)

// Request is the durable approval resource that UI, Slack, or GitHub clients can
// render and resolve.
type Request struct {
	ID               string       `json:"id"`
	Action           Action       `json:"action"`
	Decision         Decision     `json:"decision"`
	State            RequestState `json:"state"`
	CreatedAt        time.Time    `json:"createdAt"`
	UpdatedAt        time.Time    `json:"updatedAt"`
	ResolvedAt       *time.Time   `json:"resolvedAt,omitempty"`
	ResolvedBy       string       `json:"resolvedBy,omitempty"`
	ResolutionReason string       `json:"resolutionReason,omitempty"`
}

// NewRequest creates a pending durable approval request.
func NewRequest(action Action, decision Decision) Request {
	now := time.Now().UTC()
	id := strings.TrimSpace(action.ID)
	if id == "" {
		id = fmt.Sprintf("approval-%d", now.UnixNano())
	} else {
		id = strings.TrimPrefix(id, "approval-")
		id = "approval-" + safeID(id)
	}
	return Request{
		ID:        id,
		Action:    action,
		Decision:  normalizeDecision(decision),
		State:     RequestPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Store persists approval requests.
type Store interface {
	Create(ctx context.Context, request Request) (Request, error)
	Get(ctx context.Context, id string) (Request, error)
	Resolve(ctx context.Context, id string, state RequestState, resolvedBy, reason string) (Request, error)
}

// InMemoryStore is a test and local prototype approval store.
type InMemoryStore struct {
	mu       sync.Mutex
	requests map[string]Request
}

// NewInMemoryStore creates an empty in-memory approval store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{requests: map[string]Request{}}
}

// Create stores a pending approval request.
func (s *InMemoryStore) Create(ctx context.Context, request Request) (Request, error) {
	if err := ctx.Err(); err != nil {
		return Request{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if request.ID == "" {
		request = NewRequest(request.Action, request.Decision)
	}
	if _, ok := s.requests[request.ID]; ok {
		return Request{}, fmt.Errorf("approval request %q already exists", request.ID)
	}
	if request.State == "" {
		request.State = RequestPending
	}
	now := time.Now().UTC()
	if request.CreatedAt.IsZero() {
		request.CreatedAt = now
	}
	if request.UpdatedAt.IsZero() {
		request.UpdatedAt = now
	}
	s.requests[request.ID] = cloneRequest(request)
	return cloneRequest(request), nil
}

// Get loads an approval request by ID.
func (s *InMemoryStore) Get(ctx context.Context, id string) (Request, error) {
	if err := ctx.Err(); err != nil {
		return Request{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	request, ok := s.requests[id]
	if !ok {
		return Request{}, fmt.Errorf("approval request %q not found", id)
	}
	return cloneRequest(request), nil
}

// Resolve moves a pending request to a terminal state.
func (s *InMemoryStore) Resolve(ctx context.Context, id string, state RequestState, resolvedBy, reason string) (Request, error) {
	if err := ctx.Err(); err != nil {
		return Request{}, err
	}
	if state != RequestApproved && state != RequestDenied && state != RequestCanceled {
		return Request{}, fmt.Errorf("invalid approval resolution state %q", state)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	request, ok := s.requests[id]
	if !ok {
		return Request{}, fmt.Errorf("approval request %q not found", id)
	}
	if request.State != RequestPending {
		return Request{}, fmt.Errorf("approval request %q is already %s", id, request.State)
	}
	now := time.Now().UTC()
	request.State = state
	request.UpdatedAt = now
	request.ResolvedAt = &now
	request.ResolvedBy = resolvedBy
	request.ResolutionReason = reason
	s.requests[id] = request
	return cloneRequest(request), nil
}

func normalizeDecision(decision Decision) Decision {
	if decision.Kind == "" {
		decision.Kind = DecisionAllow
	}
	if decision.Risk == "" {
		decision.Risk = RiskLow
	}
	return decision
}

func maxDecision(current, candidate DecisionKind) DecisionKind {
	if decisionRank(candidate) > decisionRank(current) {
		return candidate
	}
	return current
}

func decisionRank(kind DecisionKind) int {
	switch kind {
	case DecisionDeny:
		return 3
	case DecisionNeedsHuman:
		return 2
	case DecisionAllow:
		return 1
	default:
		return 0
	}
}

func maxRisk(current, candidate Risk) Risk {
	if riskRank(candidate) > riskRank(current) {
		return candidate
	}
	return current
}

func riskRank(risk Risk) int {
	switch risk {
	case RiskCritical:
		return 4
	case RiskHigh:
		return 3
	case RiskMedium:
		return 2
	case RiskLow:
		return 1
	default:
		return 0
	}
}

func actionText(action Action) string {
	parts := []string{action.ID, action.Name, action.Kind, action.Actor, action.Subject, action.Command, action.Payload}
	parts = append(parts, action.Paths...)
	if action.Metadata != nil {
		for key, value := range action.Metadata {
			parts = append(parts, key, value)
		}
	}
	return strings.ToLower(strings.Join(parts, " "))
}

func safeID(id string) string {
	id = strings.TrimSpace(id)
	var b strings.Builder
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	if b.Len() == 0 {
		return "request"
	}
	return b.String()
}

func cloneRequest(request Request) Request {
	request.Action.Paths = append([]string(nil), request.Action.Paths...)
	if request.Action.Metadata != nil {
		metadata := make(map[string]string, len(request.Action.Metadata))
		for key, value := range request.Action.Metadata {
			metadata[key] = value
		}
		request.Action.Metadata = metadata
	}
	request.Decision.Reasons = append([]string(nil), request.Decision.Reasons...)
	request.Decision.RuleIDs = append([]string(nil), request.Decision.RuleIDs...)
	if request.ResolvedAt != nil {
		resolvedAt := *request.ResolvedAt
		request.ResolvedAt = &resolvedAt
	}
	return request
}

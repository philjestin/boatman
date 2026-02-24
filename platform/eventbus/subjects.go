package eventbus

import "fmt"

// Subject pattern constants for the NATS event bus.
// The hierarchy is: boatman.{org_id}.{team_id}.{event_type}

const (
	SubjectRunStarted      = "run.started"
	SubjectRunCompleted    = "run.completed"
	SubjectStepPrefix      = "step."
	SubjectCostRecorded    = "cost.recorded"
	SubjectBudgetAlert     = "budget.alert"
	SubjectPolicyViolation = "policy.violation"
)

// BuildSubject creates a fully qualified NATS subject.
func BuildSubject(orgID, teamID, eventType string) string {
	return fmt.Sprintf("boatman.%s.%s.%s", orgID, teamID, eventType)
}

// OrgWildcard returns a wildcard subscription for all events in an org.
// e.g., "boatman.myorg.>"
func OrgWildcard(orgID string) string {
	return fmt.Sprintf("boatman.%s.>", orgID)
}

// TeamWildcard returns a wildcard subscription for all events in a team.
// e.g., "boatman.myorg.myteam.>"
func TeamWildcard(orgID, teamID string) string {
	return fmt.Sprintf("boatman.%s.%s.>", orgID, teamID)
}

// AllEventsSubject subscribes to all boatman events.
const AllEventsSubject = "boatman.>"

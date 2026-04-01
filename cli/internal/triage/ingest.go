package triage

import (
	"regexp"
	"strings"
	"time"

	"github.com/philjestin/boatmanmode/internal/linear"
)

// filePathRe matches file paths like path/to/file.ext with at least one slash
// and a file extension.
var filePathRe = regexp.MustCompile(`[\w./-]+/[\w./-]+\.\w+`)

// ticketRefRe matches ticket references like ENG-123, FE-456.
var ticketRefRe = regexp.MustCompile(`[A-Z]+-\d+`)

// domainKeywords maps keywords found in labels, file content, or description
// text to domain names.
var domainKeywords = map[string][]string{
	"frontend": {"frontend", "react", "typescript", "component", "css", "tsx"},
	"backend":  {"backend", "rails", "ruby", "controller", "model", "service"},
	"graphql":  {"graphql", "resolver", "mutation", "query", "schema"},
	"testing":  {"test", "spec", "rspec", "jest"},
	"database": {"migration", "migrate"},
	"api":      {"api", "endpoint", "rest"},
}

// domainFileExtensions maps file extensions to domains.
var domainFileExtensions = map[string]string{
	".tsx":  "frontend",
	".ts":   "frontend",
	".jsx":  "frontend",
	".js":   "frontend",
	".css":  "frontend",
	".scss": "frontend",
	".rb":   "backend",
	".erb":  "backend",
	".graphql": "graphql",
}

// Normalize converts a linear.FullTicket into a NormalizedTicket with
// extracted signals and a staleness deadline.
func Normalize(ticket *linear.FullTicket, stalenessHours int) NormalizedTicket {
	now := time.Now()
	desc := ticket.Description

	labels := make([]string, len(ticket.Labels))
	copy(labels, ticket.Labels)

	files := extractFilePaths(desc)
	domains := extractDomains(labels, files, desc)
	deps := extractDependencies(desc)
	acPresent, acExplicit := detectAcceptanceCriteria(desc)
	hasSpec := detectDesignSpec(desc)

	summary := desc
	if len(summary) > 200 {
		summary = summary[:200]
	}
	summary = strings.TrimSpace(summary)

	return NormalizedTicket{
		TicketID:    ticket.Identifier,
		Title:       ticket.Title,
		Summary:     summary,
		Description: desc,
		IngestedAt:  now,
		StaleAfter:  now.Add(time.Duration(stalenessHours) * time.Hour),
		Signals: Signals{
			MentionsFiles:              files,
			Domains:                    domains,
			Dependencies:               deps,
			Labels:                     labels,
			AcceptanceCriteriaPresent:  acPresent,
			AcceptanceCriteriaExplicit: acExplicit,
			HasDesignSpec:              hasSpec,
			CommentCount:               len(ticket.Comments),
			LastUpdated:                ticket.UpdatedAt,
			TeamKey:                    ticket.Team,
			ProjectName:                ticket.ProjectName,
			Estimate:                   ticket.Estimate,
		},
	}
}

// NormalizeBatch converts a slice of FullTickets into NormalizedTickets.
func NormalizeBatch(tickets []linear.FullTicket, stalenessHours int) []NormalizedTicket {
	result := make([]NormalizedTicket, len(tickets))
	for i := range tickets {
		result[i] = Normalize(&tickets[i], stalenessHours)
	}
	return result
}

// extractFilePaths finds file path references in text. Matches patterns like
// path/to/file.ext, packs/something/app/models/foo.rb, next/packages/foo/bar.tsx.
// Results are deduplicated.
func extractFilePaths(text string) []string {
	matches := filePathRe.FindAllString(text, -1)
	seen := make(map[string]struct{}, len(matches))
	var result []string
	for _, m := range matches {
		if _, ok := seen[m]; !ok {
			seen[m] = struct{}{}
			result = append(result, m)
		}
	}
	return result
}

// extractDomains infers domains from labels, file extensions, and description text.
func extractDomains(labels []string, files []string, text string) []string {
	seen := make(map[string]struct{})

	// Check labels against domain keywords.
	for _, label := range labels {
		lower := strings.ToLower(label)
		for domain, keywords := range domainKeywords {
			for _, kw := range keywords {
				if strings.Contains(lower, kw) {
					seen[domain] = struct{}{}
				}
			}
		}
	}

	// Check file extensions.
	for _, f := range files {
		for ext, domain := range domainFileExtensions {
			if strings.HasSuffix(f, ext) {
				seen[domain] = struct{}{}
			}
		}
	}

	// Check description text against domain keywords.
	lower := strings.ToLower(text)
	for domain, keywords := range domainKeywords {
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				seen[domain] = struct{}{}
			}
		}
	}

	result := make([]string, 0, len(seen))
	for d := range seen {
		result = append(result, d)
	}
	return result
}

// extractDependencies extracts ticket references like ENG-123, FE-456 from text.
// Results are deduplicated.
func extractDependencies(text string) []string {
	matches := ticketRefRe.FindAllString(text, -1)
	seen := make(map[string]struct{}, len(matches))
	var result []string
	for _, m := range matches {
		if _, ok := seen[m]; !ok {
			seen[m] = struct{}{}
			result = append(result, m)
		}
	}
	return result
}

// detectAcceptanceCriteria checks whether the text contains acceptance criteria.
// Returns (present, explicit): present is true if any criteria pattern is found;
// explicit is true specifically when checkboxes or an "acceptance criteria" header
// are found.
func detectAcceptanceCriteria(text string) (present bool, explicit bool) {
	lower := strings.ToLower(text)

	hasCheckboxes := strings.Contains(text, "- [ ]") || strings.Contains(text, "- [x]")
	hasACHeader := strings.Contains(lower, "acceptance criteria")
	hasNumberedReqs := regexp.MustCompile(`(?m)^\s*\d+\.\s+\S`).MatchString(text)

	present = hasCheckboxes || hasACHeader || hasNumberedReqs
	explicit = hasCheckboxes || hasACHeader
	return
}

// detectDesignSpec returns true if the text mentions design artifacts such as
// Figma links, design specs, mockups, wireframes, or prototypes.
func detectDesignSpec(text string) bool {
	lower := strings.ToLower(text)
	designKeywords := []string{"figma", "design spec", "mockup", "wireframe", "prototype"}
	for _, kw := range designKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

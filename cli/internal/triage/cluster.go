package triage

import (
	"fmt"
	"sort"
	"strings"
)

const overlapThreshold = 2.0

// ClusterTickets groups related tickets using greedy single-linkage clustering
// based on shared signals and generates context documents for each cluster.
func ClusterTickets(tickets []NormalizedTicket, classifications []Classification) ([]Cluster, []ContextDoc) {
	// Build a map from ticketID to index for fast lookup.
	ticketIndex := make(map[string]int, len(tickets))
	for i, t := range tickets {
		ticketIndex[t.TicketID] = i
	}

	// classificationIndex maps ticketID to classification.
	classificationIndex := make(map[string]*Classification, len(classifications))
	for i := range classifications {
		classificationIndex[classifications[i].TicketID] = &classifications[i]
	}

	// clusters is a list of ticket index groups.
	type indexCluster struct {
		indices []int
	}
	var clusters []indexCluster

	for i := range tickets {
		bestCluster := -1
		bestScore := 0.0

		for ci, c := range clusters {
			// Max overlap with any member in the cluster (single-linkage).
			for _, mi := range c.indices {
				score := overlapScore(&tickets[i], &tickets[mi])
				if score > bestScore {
					bestScore = score
					bestCluster = ci
				}
			}
		}

		if bestScore >= overlapThreshold && bestCluster >= 0 {
			clusters[bestCluster].indices = append(clusters[bestCluster].indices, i)
		} else {
			clusters = append(clusters, indexCluster{indices: []int{i}})
		}
	}

	// Build Cluster and ContextDoc output.
	resultClusters := make([]Cluster, 0, len(clusters))
	contextDocs := make([]ContextDoc, 0, len(clusters))

	for ci, c := range clusters {
		clusterTickets := make([]NormalizedTicket, len(c.indices))
		ticketIDs := make([]string, len(c.indices))
		for j, idx := range c.indices {
			clusterTickets[j] = tickets[idx]
			ticketIDs[j] = tickets[idx].TicketID
		}

		// Derive cluster ID from primary domain or index.
		clusterID := deriveClusterID(ci, clusterTickets)

		// Union of repo areas.
		repoAreas := unionMentionsFiles(clusterTickets)

		rationale := buildClusterRationale(clusterTickets)

		cluster := Cluster{
			ClusterID: clusterID,
			Rationale: rationale,
			TicketIDs: ticketIDs,
			RepoAreas: repoAreas,
		}
		resultClusters = append(resultClusters, cluster)

		// Collect classifications for this cluster.
		var clusterClassifications []Classification
		for _, tid := range ticketIDs {
			if cl, ok := classificationIndex[tid]; ok {
				clusterClassifications = append(clusterClassifications, *cl)
			}
		}

		doc := generateContextDoc(cluster, clusterTickets, clusterClassifications)
		contextDocs = append(contextDocs, doc)
	}

	return resultClusters, contextDocs
}

// overlapScore computes a similarity score between two tickets based on shared signals.
func overlapScore(a, b *NormalizedTicket) float64 {
	score := 0.0

	// Shared mentionsFiles: file path prefix match scores +3.0 each.
	for _, af := range a.Signals.MentionsFiles {
		for _, bf := range b.Signals.MentionsFiles {
			if strings.HasPrefix(af, bf) || strings.HasPrefix(bf, af) {
				score += 3.0
				break
			}
		}
	}

	// Shared domains: +1.0 each.
	shared := intersect(a.Signals.Domains, b.Signals.Domains)
	score += float64(len(shared)) * 1.0

	// Shared labels: +0.5 each.
	sharedLabels := intersect(a.Signals.Labels, b.Signals.Labels)
	score += float64(len(sharedLabels)) * 0.5

	// Shared dependencies: +1.5 each.
	sharedDeps := intersect(a.Signals.Dependencies, b.Signals.Dependencies)
	score += float64(len(sharedDeps)) * 1.5

	return score
}

// deriveClusterID generates a cluster identifier from the primary domain
// or falls back to a numeric index.
func deriveClusterID(index int, tickets []NormalizedTicket) string {
	// Count domain frequency.
	domainCount := make(map[string]int)
	for _, t := range tickets {
		for _, d := range t.Signals.Domains {
			domainCount[d]++
		}
	}

	if len(domainCount) > 0 {
		// Pick the most frequent domain.
		bestDomain := ""
		bestCount := 0
		for d, c := range domainCount {
			if c > bestCount || (c == bestCount && d < bestDomain) {
				bestDomain = d
				bestCount = c
			}
		}
		return fmt.Sprintf("cluster-%s-%d", strings.ToLower(bestDomain), index+1)
	}

	return fmt.Sprintf("cluster-%d", index+1)
}

// buildClusterRationale creates a human-readable explanation of why tickets were grouped.
func buildClusterRationale(tickets []NormalizedTicket) string {
	if len(tickets) == 1 {
		return fmt.Sprintf("Standalone ticket: %s", tickets[0].TicketID)
	}

	domains := make(map[string]bool)
	files := make(map[string]bool)
	for _, t := range tickets {
		for _, d := range t.Signals.Domains {
			domains[d] = true
		}
		for _, f := range t.Signals.MentionsFiles {
			files[f] = true
		}
	}

	parts := []string{}
	if len(domains) > 0 {
		domainList := sortedKeys(domains)
		parts = append(parts, fmt.Sprintf("shared domains: %s", strings.Join(domainList, ", ")))
	}
	if len(files) > 0 {
		fileList := sortedKeys(files)
		if len(fileList) > 3 {
			fileList = fileList[:3]
			fileList = append(fileList, "...")
		}
		parts = append(parts, fmt.Sprintf("shared code areas: %s", strings.Join(fileList, ", ")))
	}

	if len(parts) == 0 {
		return fmt.Sprintf("%d tickets grouped by signal overlap", len(tickets))
	}
	return fmt.Sprintf("%d tickets grouped by %s", len(tickets), strings.Join(parts, "; "))
}

// unionMentionsFiles collects all unique mentionsFiles across a set of tickets.
func unionMentionsFiles(tickets []NormalizedTicket) []string {
	seen := make(map[string]bool)
	for _, t := range tickets {
		for _, f := range t.Signals.MentionsFiles {
			seen[f] = true
		}
	}
	return sortedKeys(seen)
}

// generateContextDoc builds a ContextDoc for a cluster of tickets.
func generateContextDoc(cluster Cluster, tickets []NormalizedTicket, classifications []Classification) ContextDoc {
	// RepoAreas already computed in the cluster.
	repoAreas := cluster.RepoAreas

	// KnownPatterns: derive from shared domains.
	domainPatterns := map[string]string{
		"frontend":  "Follow existing React component patterns",
		"backend":   "Follow existing Rails/pack patterns",
		"graphql":   "Follow existing GraphQL type and resolver patterns",
		"api":       "Follow existing API endpoint patterns",
		"migration": "Follow zero-downtime migration conventions",
		"test":      "Follow existing test patterns and conventions",
	}

	seenPatterns := make(map[string]bool)
	var knownPatterns []string
	for _, t := range tickets {
		for _, d := range t.Signals.Domains {
			lower := strings.ToLower(d)
			if pattern, ok := domainPatterns[lower]; ok && !seenPatterns[pattern] {
				seenPatterns[pattern] = true
				knownPatterns = append(knownPatterns, pattern)
			}
		}
	}
	sort.Strings(knownPatterns)

	// ValidationPlan: derive from domains.
	domainValidation := map[string]string{
		"frontend": "yarn test",
		"backend":  "bundle exec rspec",
		"graphql":  "bundle exec rspec",
		"api":      "bundle exec rspec",
		"test":     "bundle exec rspec",
	}

	seenValidation := make(map[string]bool)
	var validationPlan []string
	for _, t := range tickets {
		for _, d := range t.Signals.Domains {
			lower := strings.ToLower(d)
			if v, ok := domainValidation[lower]; ok && !seenValidation[v] {
				seenValidation[v] = true
				validationPlan = append(validationPlan, v)
			}
		}
	}
	sort.Strings(validationPlan)

	// Risks: collect uncertainAxes from classifications.
	seenRisks := make(map[string]bool)
	var risks []string
	for _, c := range classifications {
		for _, axis := range c.UncertainAxes {
			if !seenRisks[axis] {
				seenRisks[axis] = true
				risks = append(risks, axis)
			}
		}
	}
	sort.Strings(risks)

	// CostCeiling: based on highest category in cluster.
	ceiling := deriveCostCeiling(classifications)

	return ContextDoc{
		ClusterID:      cluster.ClusterID,
		Rationale:      cluster.Rationale,
		TicketIDs:      cluster.TicketIDs,
		RepoAreas:      repoAreas,
		KnownPatterns:  knownPatterns,
		ValidationPlan: validationPlan,
		Risks:          risks,
		CostCeiling:    ceiling,
	}
}

// deriveCostCeiling returns cost limits based on the highest category in the cluster.
func deriveCostCeiling(classifications []Classification) CostCeiling {
	hasDefinite := false
	for _, c := range classifications {
		if c.Category == CategoryAIDefinite {
			hasDefinite = true
			break
		}
	}

	if hasDefinite {
		// AI_DEFINITE: 500K tokens, 30 minutes.
		return CostCeiling{
			MaxTokensPerTicket:       500000,
			MaxAgentMinutesPerTicket: 30,
		}
	}

	// AI_LIKELY and others: 1M tokens, 60 minutes.
	return CostCeiling{
		MaxTokensPerTicket:       1000000,
		MaxAgentMinutesPerTicket: 60,
	}
}

// intersect returns elements present in both slices.
func intersect(a, b []string) []string {
	set := make(map[string]bool, len(a))
	for _, v := range a {
		set[v] = true
	}

	var result []string
	for _, v := range b {
		if set[v] {
			result = append(result, v)
		}
	}
	return result
}

// sortedKeys returns the keys of a map sorted alphabetically.
func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

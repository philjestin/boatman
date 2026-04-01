package triage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/linear"
	"github.com/philjestin/boatmanmode/internal/logger"
)

// PipelineOptions configures a triage pipeline run.
type PipelineOptions struct {
	// TicketIDs fetches specific tickets by identifier (e.g., "ENG-123").
	// If set, TeamKeys/States/Limit are ignored.
	TicketIDs []string

	// TeamKeys filters by team key (e.g., "ENG", "FE").
	TeamKeys []string

	// States filters by workflow state type (e.g., "backlog", "triage").
	States []string

	// Limit caps the total number of tickets fetched.
	Limit int

	// PostComments posts triage results as Linear comments.
	PostComments bool

	// DryRun skips posting comments and writing decision logs.
	DryRun bool

	// Concurrency controls parallel Claude scoring calls.
	Concurrency int

	// OutputDir is the directory for decision logs and context docs.
	OutputDir string

	// EmitEvents enables JSON event emission to stdout for desktop app integration.
	EmitEvents bool

	// GeneratePlans enables Stage 4 plan generation for AI_DEFINITE tickets.
	GeneratePlans bool

	// RepoPath is the path to the repo for plan generation and validation.
	// Required when GeneratePlans is true.
	RepoPath string
}

// Pipeline orchestrates the triage stages:
// fetch -> ingest -> score -> classify -> cluster -> log -> comment
type Pipeline struct {
	cfg          *config.Config
	linearClient *linear.Client
	scorer       *Scorer
	log          *slog.Logger
}

// NewPipeline creates a Pipeline with the given config and Linear client.
func NewPipeline(cfg *config.Config, linearClient *linear.Client) *Pipeline {
	return &Pipeline{
		cfg:          cfg,
		linearClient: linearClient,
		scorer:       NewScorer(cfg),
		log:          logger.WithComponent("triage"),
	}
}

// Run executes the full triage pipeline and returns the aggregate result.
func (p *Pipeline) Run(ctx context.Context, opts PipelineOptions) (*TriageResult, error) {
	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = p.cfg.Triage.MaxConcurrency
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = p.cfg.Triage.OutputDir
	}

	if opts.EmitEvents {
		emitTriageStarted(opts.Limit, opts.TeamKeys)
	}

	// --- Stage 0: Fetch tickets from Linear ---
	p.log.Info("fetching tickets from Linear")
	fullTickets, err := p.fetchTickets(ctx, opts)
	if err != nil {
		if opts.EmitEvents {
			emitTriageError(err)
		}
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	if len(fullTickets) == 0 {
		return &TriageResult{Stats: TriageStats{}}, nil
	}
	p.log.Info("fetched tickets", "count", len(fullTickets))
	if opts.EmitEvents {
		emitFetchComplete(len(fullTickets))
	}

	// --- Stage 1: Ingest (normalize + signal extraction) ---
	p.log.Info("ingesting tickets")
	normalized := NormalizeBatch(fullTickets, p.cfg.Triage.StalenessHours)

	// --- Stage 2a: Score (Claude rubric evaluation, concurrent) ---
	p.log.Info("scoring tickets", "concurrency", concurrency)
	if opts.EmitEvents {
		emitScoringStarted(len(normalized), concurrency)
		p.scorer.OnTicketScored = func(st ScoredTicket, index, total int) {
			emitTicketScored(st, index, total)
		}
	}
	scored := p.scorer.ScoreBatch(ctx, normalized, concurrency)

	// --- Stage 2b: Classify (deterministic decision tree) ---
	p.log.Info("classifying tickets")
	var classifications []Classification
	var totalTokens int
	var totalCost float64
	var failedCount int

	for _, st := range scored {
		if st.Err != nil {
			p.log.Warn("scoring failed, skipping ticket",
				"ticket", st.Ticket.TicketID,
				"error", st.Err)
			failedCount++
			continue
		}

		c := Classify(&st.Ticket, st.Response.RubricScores, st.Response.UncertainAxes, st.Response.Reasons)
		classifications = append(classifications, c)

		if st.Usage != nil {
			totalTokens += st.Usage.InputTokens + st.Usage.OutputTokens
			totalCost += st.Usage.TotalCostUSD
		}
	}

	if opts.EmitEvents {
		emitScoringComplete(len(classifications), failedCount)
		emitClassifying(len(classifications))
	}

	// --- Stage 3: Cluster + Context Docs ---
	p.log.Info("clustering tickets")
	if opts.EmitEvents {
		emitClustering(len(normalized))
	}
	clusters, contextDocs := ClusterTickets(normalized, classifications)

	// --- Decision Log ---
	if !opts.DryRun {
		p.log.Info("writing decision log", "dir", outputDir)
		if err := p.writeDecisionLog(outputDir, classifications, scored, clusters, contextDocs); err != nil {
			p.log.Warn("failed to write decision log", "error", err)
		}
	}

	// --- Post Linear Comments ---
	if opts.PostComments && !opts.DryRun {
		p.log.Info("posting triage comments to Linear")
		ticketUUIDs := buildTicketUUIDMap(fullTickets)
		if err := PostTriageComments(ctx, p.linearClient, classifications, ticketUUIDs); err != nil {
			p.log.Warn("failed to post some comments", "error", err)
		}
	}

	// --- Build result ---
	stats := buildStats(classifications, clusters, totalTokens, totalCost)

	result := &TriageResult{
		Tickets:         normalized,
		Classifications: classifications,
		Clusters:        clusters,
		ContextDocs:     contextDocs,
		Stats:           stats,
	}

	if opts.EmitEvents {
		emitTriageComplete(result)
	}

	return result, nil
}

// fetchTickets retrieves tickets from Linear based on pipeline options.
func (p *Pipeline) fetchTickets(ctx context.Context, opts PipelineOptions) ([]linear.FullTicket, error) {
	if len(opts.TicketIDs) > 0 {
		return p.fetchByIDs(ctx, opts.TicketIDs)
	}

	listOpts := linear.ListOptions{
		TeamKeys: opts.TeamKeys,
		States:   opts.States,
		Limit:    opts.Limit,
	}
	return p.linearClient.ListTickets(ctx, listOpts)
}

// fetchByIDs fetches individual tickets by identifier.
func (p *Pipeline) fetchByIDs(ctx context.Context, ids []string) ([]linear.FullTicket, error) {
	var tickets []linear.FullTicket
	for _, id := range ids {
		t, err := p.linearClient.GetFullTicket(ctx, id)
		if err != nil {
			p.log.Warn("failed to fetch ticket", "id", id, "error", err)
			continue
		}
		tickets = append(tickets, *t)
	}
	return tickets, nil
}

// writeDecisionLog persists classification and cluster decisions.
func (p *Pipeline) writeDecisionLog(outputDir string, classifications []Classification, scored []ScoredTicket, clusters []Cluster, contextDocs []ContextDoc) error {
	dl, err := NewDecisionLog(outputDir)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	// Log each classification.
	for _, c := range classifications {
		details, _ := json.Marshal(c)
		entry := DecisionLogEntry{
			TicketID:  c.TicketID,
			Stage:     StageScore,
			Verdict:   string(c.Category),
			Agent:     "triage-pipeline",
			Rationale: classificationRationale(c),
			Timestamp: now,
			Details:   details,
		}

		// Find matching scored ticket for token/cost info.
		for _, st := range scored {
			if st.Ticket.TicketID == c.TicketID && st.Usage != nil {
				entry.TokensUsed = st.Usage.InputTokens + st.Usage.OutputTokens
				entry.CostUSD = st.Usage.TotalCostUSD
				break
			}
		}

		if err := dl.Append(entry); err != nil {
			p.log.Warn("failed to log decision", "ticket", c.TicketID, "error", err)
		}
	}

	// Log cluster decisions.
	for _, cluster := range clusters {
		details, _ := json.Marshal(cluster)
		entry := DecisionLogEntry{
			TicketID:  cluster.ClusterID,
			Stage:     StageCluster,
			Verdict:   fmt.Sprintf("cluster:%d-tickets", len(cluster.TicketIDs)),
			Agent:     "triage-pipeline",
			Rationale: cluster.Rationale,
			Timestamp: now,
			Details:   details,
		}
		if err := dl.Append(entry); err != nil {
			p.log.Warn("failed to log cluster", "cluster", cluster.ClusterID, "error", err)
		}
	}

	// Write context docs.
	for _, doc := range contextDocs {
		if err := dl.WriteContextDoc(doc); err != nil {
			p.log.Warn("failed to write context doc", "cluster", doc.ClusterID, "error", err)
		}
	}

	return nil
}

// buildTicketUUIDMap creates a mapping from human-readable identifiers to UUIDs.
func buildTicketUUIDMap(tickets []linear.FullTicket) map[string]string {
	m := make(map[string]string, len(tickets))
	for _, t := range tickets {
		m[t.Identifier] = t.ID
	}
	return m
}

// classificationRationale generates a one-line summary of the classification.
func classificationRationale(c Classification) string {
	if len(c.HardStops) > 0 {
		return fmt.Sprintf("hard stops: %v", c.HardStops)
	}
	for _, g := range c.GateResults {
		if !g.Passed {
			return fmt.Sprintf("gate failed: %s", g.Gate)
		}
	}
	return fmt.Sprintf("classified as %s", c.Category)
}

// buildStats computes aggregate statistics from the pipeline run.
func buildStats(classifications []Classification, clusters []Cluster, totalTokens int, totalCost float64) TriageStats {
	stats := TriageStats{
		TotalTickets:    len(classifications),
		ClusterCount:    len(clusters),
		TotalTokensUsed: totalTokens,
		TotalCostUSD:    totalCost,
	}

	for _, c := range classifications {
		switch c.Category {
		case CategoryAIDefinite:
			stats.AIDefiniteCount++
		case CategoryAILikely:
			stats.AILikelyCount++
		case CategoryHumanReviewRequired:
			stats.HumanReviewCount++
		case CategoryHumanOnly:
			stats.HumanOnlyCount++
		}
	}

	return stats
}

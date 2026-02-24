package team

import (
	"context"
	"strings"

	"github.com/philjestin/boatman-ecosystem/harness/cost"
)

// Aggregator combines multiple agent results into a single result.
type Aggregator interface {
	Aggregate(ctx context.Context, results []Result) (*Result, error)
}

// ConcatAggregator concatenates outputs, merges FilesChanged, and sums Usage.
type ConcatAggregator struct{}

// Aggregate concatenates all outputs separated by newlines, merges file lists,
// sums usage, and preserves individual results as Children.
func (ConcatAggregator) Aggregate(_ context.Context, results []Result) (*Result, error) {
	if len(results) == 0 {
		return &Result{}, nil
	}

	var (
		outputs []string
		usage   cost.Usage
		files   []string
		data    = make(map[string]any)
	)

	for _, r := range results {
		if r.Output != "" {
			outputs = append(outputs, r.Output)
		}
		usage = usage.Add(r.Usage)
		files = append(files, r.FilesChanged...)
		for k, v := range r.Data {
			data[k] = v
		}
	}

	return &Result{
		Output:       strings.Join(outputs, "\n"),
		Data:         data,
		Usage:        usage,
		FilesChanged: dedupStrings(files),
		Children:     results,
	}, nil
}

// FirstResultAggregator returns the first non-error result.
type FirstResultAggregator struct{}

// Aggregate returns the first result without an error. If all results have
// errors, the first result is returned as-is.
func (FirstResultAggregator) Aggregate(_ context.Context, results []Result) (*Result, error) {
	if len(results) == 0 {
		return &Result{}, nil
	}
	for i := range results {
		if results[i].Error == nil {
			r := results[i]
			r.Children = results
			return &r, nil
		}
	}
	// All errored — return first.
	r := results[0]
	r.Children = results
	return &r, r.Error
}

// BestResultAggregator selects the highest-scored result using a caller-provided
// scoring function.
type BestResultAggregator struct {
	Score func(*Result) float64
}

// Aggregate returns the result with the highest score. All results are
// preserved as Children.
func (b BestResultAggregator) Aggregate(_ context.Context, results []Result) (*Result, error) {
	if len(results) == 0 {
		return &Result{}, nil
	}

	bestIdx := 0
	bestScore := b.Score(&results[0])
	for i := 1; i < len(results); i++ {
		if s := b.Score(&results[i]); s > bestScore {
			bestIdx = i
			bestScore = s
		}
	}

	r := results[bestIdx]
	r.Children = results
	return &r, nil
}

// dedupStrings returns unique strings preserving order.
func dedupStrings(ss []string) []string {
	if len(ss) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ss))
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

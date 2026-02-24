package team

import (
	"context"
	"errors"
	"testing"

	"github.com/philjestin/boatman-ecosystem/harness/cost"
)

func TestConcatAggregator(t *testing.T) {
	results := []Result{
		{AgentName: "a1", Output: "hello", FilesChanged: []string{"a.go", "b.go"}, Usage: cost.Usage{InputTokens: 100}},
		{AgentName: "a2", Output: "world", FilesChanged: []string{"b.go", "c.go"}, Usage: cost.Usage{InputTokens: 200}},
	}

	agg := ConcatAggregator{}
	r, err := agg.Aggregate(context.Background(), results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Output != "hello\nworld" {
		t.Fatalf("expected 'hello\\nworld', got %q", r.Output)
	}
	if len(r.FilesChanged) != 3 {
		t.Fatalf("expected 3 unique files, got %d: %v", len(r.FilesChanged), r.FilesChanged)
	}
	if r.Usage.InputTokens != 300 {
		t.Fatalf("expected 300 input tokens, got %d", r.Usage.InputTokens)
	}
	if len(r.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(r.Children))
	}
}

func TestConcatAggregator_Empty(t *testing.T) {
	r, err := ConcatAggregator{}.Aggregate(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Output != "" {
		t.Fatalf("expected empty output, got %q", r.Output)
	}
}

func TestConcatAggregator_MergesData(t *testing.T) {
	results := []Result{
		{AgentName: "a1", Data: map[string]any{"key1": "val1"}},
		{AgentName: "a2", Data: map[string]any{"key2": "val2"}},
	}

	r, err := ConcatAggregator{}.Aggregate(context.Background(), results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Data["key1"] != "val1" || r.Data["key2"] != "val2" {
		t.Fatalf("expected merged data, got %v", r.Data)
	}
}

func TestFirstResultAggregator(t *testing.T) {
	results := []Result{
		{AgentName: "a1", Error: errors.New("fail")},
		{AgentName: "a2", Output: "success"},
		{AgentName: "a3", Output: "also-success"},
	}

	r, err := FirstResultAggregator{}.Aggregate(context.Background(), results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.AgentName != "a2" {
		t.Fatalf("expected a2, got %s", r.AgentName)
	}
	if r.Output != "success" {
		t.Fatalf("expected 'success', got %q", r.Output)
	}
	if len(r.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(r.Children))
	}
}

func TestFirstResultAggregator_AllErrors(t *testing.T) {
	results := []Result{
		{AgentName: "a1", Error: errors.New("fail1")},
		{AgentName: "a2", Error: errors.New("fail2")},
	}

	r, err := FirstResultAggregator{}.Aggregate(context.Background(), results)
	if err == nil {
		t.Fatal("expected error when all results error")
	}
	if r.AgentName != "a1" {
		t.Fatalf("expected first result a1, got %s", r.AgentName)
	}
}

func TestFirstResultAggregator_Empty(t *testing.T) {
	r, err := FirstResultAggregator{}.Aggregate(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Output != "" {
		t.Fatalf("expected empty output, got %q", r.Output)
	}
}

func TestBestResultAggregator(t *testing.T) {
	results := []Result{
		{AgentName: "a1", Output: "low", Data: map[string]any{"score": 1.0}},
		{AgentName: "a2", Output: "high", Data: map[string]any{"score": 9.0}},
		{AgentName: "a3", Output: "mid", Data: map[string]any{"score": 5.0}},
	}

	agg := BestResultAggregator{
		Score: func(r *Result) float64 {
			if s, ok := r.Data["score"].(float64); ok {
				return s
			}
			return 0
		},
	}

	r, err := agg.Aggregate(context.Background(), results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.AgentName != "a2" {
		t.Fatalf("expected a2 (highest score), got %s", r.AgentName)
	}
	if len(r.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(r.Children))
	}
}

func TestBestResultAggregator_Empty(t *testing.T) {
	agg := BestResultAggregator{Score: func(*Result) float64 { return 0 }}
	r, err := agg.Aggregate(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Output != "" {
		t.Fatalf("expected empty output, got %q", r.Output)
	}
}

func TestDedupStrings(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{nil, nil},
		{[]string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{[]string{"x"}, []string{"x"}},
	}

	for _, tt := range tests {
		result := dedupStrings(tt.input)
		if len(result) != len(tt.expected) {
			t.Fatalf("dedupStrings(%v): expected %v, got %v", tt.input, tt.expected, result)
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Fatalf("dedupStrings(%v)[%d]: expected %s, got %s", tt.input, i, tt.expected[i], result[i])
			}
		}
	}
}

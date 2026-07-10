package brain

import (
	"path/filepath"
	"testing"
)

func TestCollectorBuildsProjectKnowledgeGraph(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)
	projectPath := t.TempDir()

	collector, err := NewCollector(projectPath)
	if err != nil {
		t.Fatalf("NewCollector: %v", err)
	}

	collector.OnTaskContext("ENG-456", "Improve slow GraphQL query", []string{
		"packs/graphql/app/resolvers/recruit_job_details.rb",
	})
	collector.OnTaskExecution("ENG-456", "Improve slow GraphQL query", []string{
		"packs/graphql/app/resolvers/recruit_job_details.rb",
	})
	collector.OnReviewFailure([]string{"N+1 query still possible"}, []string{
		"packs/graphql/app/resolvers/recruit_job_details.rb",
	})
	collector.OnRefactorIteration(2, []string{"Reviewer requested batch loading"}, []string{
		"packs/graphql/app/resolvers/recruit_job_details.rb",
	})
	collector.OnBrainsDistilled([]DistillResult{
		{
			Domain:     "graphql",
			BrainID:    "auto-graphql",
			Path:       filepath.Join(projectPath, ".boatman", "brains", "graphql.yaml"),
			MemoryPath: filepath.Join(projectPath, ".boatman", "memory", "domains", "graphql.md"),
			Signals:    2,
		},
	})

	if err := collector.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	graph, err := LoadKnowledgeGraph(projectPath)
	if err != nil {
		t.Fatalf("LoadKnowledgeGraph: %v", err)
	}
	assertNode(t, graph, "task:eng-456", "task")
	assertNode(t, graph, "domain:graphql", "domain")
	assertNode(t, graph, "signal:review-failure-graphql", "signal")
	assertNode(t, graph, "signal:refactor-loop-graphql", "signal")
	assertNode(t, graph, "brain:auto-graphql", "brain")
	assertEdge(t, graph, "task:eng-456", "signal:review-failure-graphql", "emitted_signal")
	assertEdge(t, graph, "task:eng-456", "signal:refactor-loop-graphql", "emitted_signal")
	assertEdge(t, graph, "domain:graphql", "brain:auto-graphql", "distilled_into")
}

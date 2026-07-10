package integrations

import "testing"

func TestDefaultCatalogListsSortedIntegrations(t *testing.T) {
	items := DefaultCatalog().List()
	if len(items) < 4 {
		t.Fatalf("default catalog returned %d integrations, want at least 4", len(items))
	}
	for i := 1; i < len(items); i++ {
		if items[i-1].Name > items[i].Name {
			t.Fatalf("items are not sorted: %#v", items)
		}
	}
}

func TestMCPRefsReturnsKnownRefsAndDedupes(t *testing.T) {
	refs, err := DefaultCatalog().MCPRefs("Datadog", "linear", "datadog")
	if err != nil {
		t.Fatalf("MCPRefs error: %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("refs = %#v, want two deduped refs", refs)
	}
	if refs[0].Label != "datadog" || refs[0].Command != "npx" || len(refs[0].Args) == 0 {
		t.Fatalf("datadog ref = %#v, want local MCP descriptor", refs[0])
	}
	if refs[1].Label != "linear" || refs[1].Env["LINEAR_API_KEY"] != "" {
		t.Fatalf("linear ref = %#v, want env placeholder", refs[1])
	}
}

func TestMCPRefsRejectsUnknownIntegration(t *testing.T) {
	_, err := DefaultCatalog().MCPRefs("not-real")
	if err == nil {
		t.Fatal("MCPRefs should fail for unknown integrations")
	}
}

func TestKnownMCPRefsSkipsUnknownIntegrations(t *testing.T) {
	refs := DefaultCatalog().KnownMCPRefs("custom-okta", "datadog")
	if len(refs) != 1 || refs[0].Label != "datadog" {
		t.Fatalf("refs = %#v, want only known datadog ref", refs)
	}
}

func TestMCPRefReturnsDefensiveCopies(t *testing.T) {
	item, ok := DefaultCatalog().Lookup("slack")
	if !ok {
		t.Fatal("slack integration should exist")
	}
	ref, ok := item.MCPRef()
	if !ok {
		t.Fatal("slack should expose MCP")
	}
	ref.Args[0] = "changed"
	ref.Env["SLACK_BOT_TOKEN"] = "changed"

	next, _ := item.MCPRef()
	if next.Args[0] == "changed" || next.Env["SLACK_BOT_TOKEN"] == "changed" {
		t.Fatalf("MCPRef should return defensive copies, got %#v", next)
	}
}

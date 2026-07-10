package agentruntime

import (
	"encoding/json"
	"testing"
)

func TestValidateOutputSchema(t *testing.T) {
	err := ValidateOutputSchema(&OutputSchema{
		Name:   "score",
		Strict: true,
		Schema: json.RawMessage(`{"type":"object","properties":{"ok":{"type":"boolean"}}}`),
	})
	if err != nil {
		t.Fatalf("ValidateOutputSchema returned error: %v", err)
	}
}

func TestValidateOutputSchemaRejectsMissingName(t *testing.T) {
	err := ValidateOutputSchema(&OutputSchema{
		Schema: json.RawMessage(`{"type":"object"}`),
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestValidateOutputSchemaRejectsInvalidJSON(t *testing.T) {
	err := ValidateOutputSchema(&OutputSchema{
		Name:   "bad",
		Schema: json.RawMessage(`{"type":`),
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestValidateOutputSchemaRejectsMissingType(t *testing.T) {
	err := ValidateOutputSchema(&OutputSchema{
		Name:   "bad",
		Schema: json.RawMessage(`{"properties":{}}`),
	})
	if err == nil {
		t.Fatal("expected error for missing type")
	}
}

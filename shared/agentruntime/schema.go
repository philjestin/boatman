package agentruntime

import (
	"encoding/json"
	"fmt"
)

// ValidateOutputSchema checks that a structured-output request is well-formed.
// It deliberately does not implement full JSON Schema validation; provider
// adapters should pass the schema to backends that enforce it natively when
// available.
func ValidateOutputSchema(schema *OutputSchema) error {
	if schema == nil {
		return fmt.Errorf("output schema is required")
	}
	if schema.Name == "" {
		return fmt.Errorf("output schema name is required")
	}
	if len(schema.Schema) == 0 {
		return fmt.Errorf("output schema %q body is required", schema.Name)
	}
	if !json.Valid(schema.Schema) {
		return fmt.Errorf("output schema %q is not valid JSON", schema.Name)
	}

	var decoded map[string]any
	if err := json.Unmarshal(schema.Schema, &decoded); err != nil {
		return fmt.Errorf("output schema %q cannot be decoded: %w", schema.Name, err)
	}
	schemaType, ok := decoded["type"].(string)
	if !ok || schemaType == "" {
		return fmt.Errorf("output schema %q must declare a JSON Schema type", schema.Name)
	}
	return nil
}

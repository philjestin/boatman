package agent

import (
	"os"
	"strings"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/integrations"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/toolbroker"

	runtimeproviders "boatman/agent/providers"
)

func (s *Session) buildRuntimeRequest(prompt string, authConfig AuthConfig) agentruntime.RunRequest {
	s.mu.RLock()
	model := s.Model
	effort := s.ReasoningEffort
	systemPrompt := s.systemPrompt
	conversationID := s.conversationID
	mode := s.Mode
	modeConfig := s.ModeConfig
	messageCount := len(s.Messages)
	s.mu.RUnlock()

	role := agentruntime.RoleChat
	profile := "desktop-chat"
	actualPrompt := prompt
	approvalPolicy := approvalPolicyForMode(authConfig.ApprovalMode)
	tools := toolRefsForApproval(authConfig.ApprovalMode)
	var mcpServers []agentruntime.MCPServerRef

	if mode == "firefighter" {
		mcpNames := mcpNamesFromModeConfig(modeConfig)
		role = agentruntime.RoleFirefight
		profile = "desktop-firefighter"
		approvalPolicy = agentruntime.ApprovalFullAuto
		tools = nil
		mcpServers = integrations.DefaultCatalog().KnownMCPRefs(mcpNames...)
		if messageCount <= 1 {
			scope, _ := modeConfig["scope"].(string)
			actualPrompt = GetFirefighterPrompt(scope, mcpNames...) + "\n\n" + prompt
		}
	}
	if mode == "routine" {
		role = agentruntime.RoleRoutine
		profile = stringFromModeConfig(modeConfig, "profile", "desktop-routine")
		mcpServers = mcpRefsFromModeConfig(modeConfig)
	}

	metadata := map[string]string{
		"outputFormat": "stream-json",
		"verbose":      "true",
		"approvalMode": authConfig.ApprovalMode,
		"authMethod":   authConfig.Method,
		"phaseId":      profile,
	}
	if conversationID != "" {
		metadata["conversationId"] = conversationID
	}
	if authConfig.APIKey != "" {
		metadata["anthropicApiKey"] = authConfig.APIKey
	}
	if authConfig.GCPProjectID != "" {
		metadata["gcpProjectId"] = authConfig.GCPProjectID
	}
	if authConfig.GCPRegion != "" {
		metadata["gcpRegion"] = authConfig.GCPRegion
	}

	return agentruntime.RunRequest{
		RunID:        s.ID,
		Role:         role,
		Profile:      profile,
		Provider:     desktopProvider(modeConfig),
		Model:        model,
		WorkDir:      s.ProjectPath,
		Instructions: systemPrompt,
		Messages: []agentruntime.Message{
			{Role: "user", Content: actualPrompt},
		},
		Tools:          tools,
		MCPServers:     mcpServers,
		ApprovalPolicy: approvalPolicy,
		Reasoning: &agentruntime.ReasoningOptions{
			Effort: effort,
		},
		Metadata: metadata,
	}
}

func desktopProvider(modeConfig map[string]interface{}) string {
	if modeConfig != nil {
		if provider, ok := modeConfig["provider"].(string); ok {
			if provider = strings.TrimSpace(provider); provider != "" {
				return provider
			}
		}
	}
	if provider := strings.TrimSpace(os.Getenv("BOATMAN_DESKTOP_PROVIDER")); provider != "" {
		return provider
	}
	return runtimeproviders.DefaultProvider
}

func approvalPolicyForMode(mode string) agentruntime.ApprovalPolicy {
	switch mode {
	case "auto-edit":
		return agentruntime.ApprovalAutoEdit
	case "full-auto":
		return agentruntime.ApprovalFullAuto
	default:
		return agentruntime.ApprovalSuggest
	}
}

func toolRefsForApproval(mode string) []agentruntime.ToolRef {
	if mode != "auto-edit" {
		return nil
	}
	return toolbroker.AutoEditRefs()
}

func mcpNamesFromModeConfig(modeConfig map[string]interface{}) []string {
	raw, ok := modeConfig["mcpServers"]
	if !ok {
		return nil
	}
	if names, ok := raw.([]string); ok {
		return names
	}
	if values, ok := raw.([]interface{}); ok {
		names := make([]string, 0, len(values))
		for _, value := range values {
			if name, ok := value.(string); ok {
				names = append(names, name)
			}
		}
		return names
	}
	return nil
}

func stringFromModeConfig(modeConfig map[string]interface{}, key, fallback string) string {
	if modeConfig != nil {
		if value, ok := modeConfig[key].(string); ok {
			if value = strings.TrimSpace(value); value != "" {
				return value
			}
		}
	}
	return fallback
}

func mcpRefsFromModeConfig(modeConfig map[string]interface{}) []agentruntime.MCPServerRef {
	if modeConfig == nil {
		return nil
	}
	raw, ok := modeConfig["mcpServers"]
	if !ok {
		return nil
	}
	switch refs := raw.(type) {
	case []agentruntime.MCPServerRef:
		return refs
	case []interface{}:
		out := make([]agentruntime.MCPServerRef, 0, len(refs))
		for _, value := range refs {
			if ref, ok := mcpRefFromAny(value); ok {
				out = append(out, ref)
			}
		}
		return out
	case []map[string]interface{}:
		out := make([]agentruntime.MCPServerRef, 0, len(refs))
		for _, value := range refs {
			if ref, ok := mcpRefFromMap(value); ok {
				out = append(out, ref)
			}
		}
		return out
	default:
		return nil
	}
}

func mcpRefFromAny(value interface{}) (agentruntime.MCPServerRef, bool) {
	switch typed := value.(type) {
	case agentruntime.MCPServerRef:
		return typed, strings.TrimSpace(typed.Label) != ""
	case map[string]interface{}:
		return mcpRefFromMap(typed)
	default:
		return agentruntime.MCPServerRef{}, false
	}
}

func mcpRefFromMap(value map[string]interface{}) (agentruntime.MCPServerRef, bool) {
	ref := agentruntime.MCPServerRef{
		Label:       stringMapValue(value, "label"),
		Command:     stringMapValue(value, "command"),
		URL:         stringMapValue(value, "url"),
		Description: stringMapValue(value, "description"),
		Args:        stringSliceMapValue(value, "args"),
		Env:         stringMapMapValue(value, "env"),
	}
	return ref, strings.TrimSpace(ref.Label) != ""
}

func stringMapValue(value map[string]interface{}, key string) string {
	raw, ok := value[key]
	if !ok {
		return ""
	}
	text, _ := raw.(string)
	return text
}

func stringSliceMapValue(value map[string]interface{}, key string) []string {
	raw, ok := value[key]
	if !ok {
		return nil
	}
	if items, ok := raw.([]string); ok {
		return items
	}
	items, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			out = append(out, text)
		}
	}
	return out
}

func stringMapMapValue(value map[string]interface{}, key string) map[string]string {
	raw, ok := value[key]
	if !ok {
		return nil
	}
	if items, ok := raw.(map[string]string); ok {
		return items
	}
	items, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}
	out := make(map[string]string, len(items))
	for itemKey, itemValue := range items {
		if text, ok := itemValue.(string); ok {
			out[itemKey] = text
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

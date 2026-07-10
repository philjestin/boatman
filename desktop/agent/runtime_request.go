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

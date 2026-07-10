// Package openairesponses adapts OpenAI's Responses API to the provider-neutral
// agentruntime.Provider contract.
package openairesponses

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/toolbroker"
)

const providerName = "openai-responses"

// Provider calls the OpenAI Responses API through the shared runtime contract.
type Provider struct {
	apiKey       string
	baseURL      string
	defaultModel string
	organization string
	project      string
	httpClient   *http.Client
	broker       *toolbroker.Broker
	maxToolTurns int
}

// Option configures a Provider.
type Option func(*Provider)

// WithAPIKey sets the OpenAI API key used for requests.
func WithAPIKey(apiKey string) Option {
	return func(p *Provider) {
		p.apiKey = apiKey
	}
}

// WithBaseURL sets the API base URL. It should usually end in /v1.
func WithBaseURL(baseURL string) Option {
	return func(p *Provider) {
		p.baseURL = baseURL
	}
}

// WithDefaultModel sets the model used when RunRequest.Model is empty.
func WithDefaultModel(model string) Option {
	return func(p *Provider) {
		p.defaultModel = model
	}
}

// WithHTTPClient sets the HTTP client used for API calls.
func WithHTTPClient(client *http.Client) Option {
	return func(p *Provider) {
		if client != nil {
			p.httpClient = client
		}
	}
}

// WithToolBroker enables local function-tool execution through Boatman's
// provider-neutral broker.
func WithToolBroker(broker *toolbroker.Broker) Option {
	return func(p *Provider) {
		p.broker = broker
	}
}

// WithMaxToolTurns caps the number of model/tool follow-up turns.
func WithMaxToolTurns(turns int) Option {
	return func(p *Provider) {
		if turns > 0 {
			p.maxToolTurns = turns
		}
	}
}

// New creates an OpenAI Responses provider adapter.
func New(opts ...Option) *Provider {
	p := &Provider{
		apiKey:       os.Getenv("OPENAI_API_KEY"),
		baseURL:      envOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		defaultModel: os.Getenv("OPENAI_MODEL"),
		organization: os.Getenv("OPENAI_ORG_ID"),
		project:      os.Getenv("OPENAI_PROJECT_ID"),
		httpClient:   &http.Client{Timeout: 10 * time.Minute},
		maxToolTurns: 4,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return providerName
}

// Capabilities reports the features this adapter currently exposes.
func (p *Provider) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return agentruntime.Capabilities{
		Provider:              providerName,
		SupportsStreaming:     false,
		SupportsBackground:    false,
		SupportsResume:        false,
		SupportsToolCalls:     true,
		SupportsMCP:           true,
		SupportsApprovals:     true,
		SupportsStructuredOut: true,
		SupportsArtifacts:     false,
		SupportsUsage:         true,
		SupportsVision:        false,
		SupportsAudio:         false,
		SupportsComputerUse:   false,
		Experimental:          []string{"responses-api", "hosted-tools", "remote-mcp", "reasoning"},
	}, nil
}

// StartRun starts a one-shot Responses API request and returns normalized events.
func (p *Provider) StartRun(ctx context.Context, req agentruntime.RunRequest) (agentruntime.EventStream, error) {
	if req.OutputSchema != nil {
		if err := agentruntime.ValidateOutputSchema(req.OutputSchema); err != nil {
			return nil, err
		}
	}

	events := make(chan agentruntime.Event, 32)
	runID := req.RunID
	if runID == "" {
		runID = fmt.Sprintf("run-%d", time.Now().UnixNano())
	}
	phaseID := phaseID(req)

	go func() {
		defer close(events)

		emit := func(event agentruntime.Event) bool {
			event.RunID = runID
			if event.PhaseID == "" {
				event.PhaseID = phaseID
			}
			event.Provider = providerName
			event.Model = modelFor(req, p.defaultModel)
			event.Role = req.Role
			select {
			case <-ctx.Done():
				return false
			case events <- event:
				return true
			}
		}

		started := agentruntime.NewEvent(agentruntime.EventRunStarted)
		started.Status = agentruntime.StatusStarted
		started.Name = req.Profile
		emit(started)

		phaseStarted := agentruntime.NewEvent(agentruntime.EventPhaseStarted)
		phaseStarted.Status = agentruntime.StatusStarted
		phaseStarted.Name = string(req.Role)
		emit(phaseStarted)

		response, err := p.runResponses(ctx, req, emit)
		if err != nil {
			failed := agentruntime.NewEvent(agentruntime.EventPhaseCompleted)
			failed.Status = agentruntime.StatusFailed
			failed.Message = err.Error()
			emit(failed)

			runFailed := agentruntime.NewEvent(agentruntime.EventRunFailed)
			runFailed.Status = agentruntime.StatusFailed
			runFailed.Message = err.Error()
			emit(runFailed)
			return
		}

		if text := response.text(); text != "" {
			message := agentruntime.NewEvent(agentruntime.EventMessageCompleted)
			message.Message = text
			emit(message)
		}

		completed := agentruntime.NewEvent(agentruntime.EventPhaseCompleted)
		completed.Status = agentruntime.StatusSucceeded
		emit(completed)

		runCompleted := agentruntime.NewEvent(agentruntime.EventRunCompleted)
		runCompleted.Status = agentruntime.StatusSucceeded
		emit(runCompleted)
	}()

	return events, nil
}

// ResumeRun is not implemented for the current one-shot adapter.
func (p *Provider) ResumeRun(context.Context, string, agentruntime.RunInput) (agentruntime.EventStream, error) {
	return nil, fmt.Errorf("%s does not support runtime resume yet", providerName)
}

// CancelRun is handled by canceling the context passed to StartRun.
func (p *Provider) CancelRun(context.Context, string) error {
	return nil
}

func (p *Provider) runResponses(ctx context.Context, req agentruntime.RunRequest, emit func(agentruntime.Event) bool) (openAIResponse, error) {
	payload, err := p.requestPayload(req)
	if err != nil {
		return openAIResponse{}, err
	}
	input, _ := payload["input"].([]any)

	var response openAIResponse
	for turn := 0; turn <= p.maxToolTurns; turn++ {
		raw, next, err := p.createResponse(ctx, payload)
		if len(raw) > 0 {
			rawEvent := agentruntime.NewEvent(agentruntime.EventProviderRaw)
			rawEvent.Raw = json.RawMessage(raw)
			emit(rawEvent)
		}
		if usage := next.runtimeUsage(); usage != nil {
			usageEvent := agentruntime.NewEvent(agentruntime.EventUsageUpdated)
			usageEvent.Usage = usage
			emit(usageEvent)
		}
		if err != nil {
			return next, err
		}
		response = next

		calls := response.functionCalls()
		if len(calls) == 0 {
			return response, nil
		}
		if p.broker == nil {
			return response, errors.New("openai-responses received function calls but no tool broker is configured")
		}
		if turn == p.maxToolTurns {
			return response, fmt.Errorf("openai-responses exceeded max tool turns (%d)", p.maxToolTurns)
		}

		for _, item := range response.Output {
			input = append(input, item.asInput())
		}
		for _, call := range calls {
			toolCall := agentruntime.NewEvent(agentruntime.EventToolCall)
			toolCall.Tool = &agentruntime.ToolEvent{
				ID:    call.CallID,
				Name:  call.Name,
				Input: json.RawMessage(call.Arguments),
			}
			emit(toolCall)

			result, toolErr := p.broker.Invoke(ctx, toolbroker.Invocation{
				ID:             call.CallID,
				Name:           call.Name,
				WorkDir:        req.WorkDir,
				Input:          json.RawMessage(call.Arguments),
				ApprovalPolicy: req.ApprovalPolicy,
			})

			toolResult := agentruntime.NewEvent(agentruntime.EventToolResult)
			toolResult.Tool = &agentruntime.ToolEvent{
				ID:      call.CallID,
				Name:    call.Name,
				Output:  result.Output,
				IsError: toolErr != nil,
			}
			emit(toolResult)

			input = append(input, map[string]any{
				"type":    "function_call_output",
				"call_id": call.CallID,
				"output":  functionOutput(result, toolErr),
			})
		}
		payload["input"] = input
	}

	return response, nil
}

func (p *Provider) createResponse(ctx context.Context, payload map[string]any) ([]byte, openAIResponse, error) {
	if strings.TrimSpace(p.apiKey) == "" {
		return nil, openAIResponse{}, errors.New("OPENAI_API_KEY is required for openai-responses")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, openAIResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, responseEndpoint(p.baseURL), bytes.NewReader(body))
	if err != nil {
		return nil, openAIResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(p.apiKey))
	if p.organization != "" {
		httpReq.Header.Set("OpenAI-Organization", p.organization)
	}
	if p.project != "" {
		httpReq.Header.Set("OpenAI-Project", p.project)
	}

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, openAIResponse{}, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, openAIResponse{}, err
	}

	var parsed openAIResponse
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return raw, openAIResponse{}, fmt.Errorf("decode OpenAI response: %w", err)
		}
	}
	if resp.StatusCode >= 300 {
		return raw, parsed, fmt.Errorf("OpenAI Responses API returned %s: %s", resp.Status, parsed.errorMessage())
	}
	if parsed.Error != nil {
		return raw, parsed, errors.New(parsed.errorMessage())
	}
	return raw, parsed, nil
}

func (p *Provider) requestPayload(req agentruntime.RunRequest) (map[string]any, error) {
	if req.Background {
		return nil, errors.New("openai-responses background runs are not wired to Boatman resume/polling yet")
	}
	model := modelFor(req, p.defaultModel)
	if model == "" {
		return nil, errors.New("openai-responses requires RunRequest.Model or OPENAI_MODEL")
	}
	tools, err := openAITools(req.Tools, req.MCPServers, req.ApprovalPolicy, p.broker != nil)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"model": model,
		"input": inputMessages(req),
		"store": false,
	}
	if len(tools) > 0 {
		payload["tools"] = tools
	}
	if req.Reasoning != nil {
		reasoning := map[string]any{}
		if req.Reasoning.Effort != "" {
			reasoning["effort"] = req.Reasoning.Effort
		}
		if len(reasoning) > 0 {
			payload["reasoning"] = reasoning
		}
		if req.Reasoning.MaxTokens > 0 {
			payload["max_output_tokens"] = req.Reasoning.MaxTokens
		}
	}
	if req.OutputSchema != nil {
		var schema any
		if err := json.Unmarshal(req.OutputSchema.Schema, &schema); err != nil {
			return nil, fmt.Errorf("decode output schema: %w", err)
		}
		format := map[string]any{
			"type":   "json_schema",
			"name":   req.OutputSchema.Name,
			"schema": schema,
			"strict": req.OutputSchema.Strict,
		}
		if req.OutputSchema.Description != "" {
			format["description"] = req.OutputSchema.Description
		}
		payload["text"] = map[string]any{"format": format}
	}
	return payload, nil
}

func inputMessages(req agentruntime.RunRequest) []any {
	input := make([]any, 0, len(req.Messages)+1)
	if strings.TrimSpace(req.Instructions) != "" {
		input = append(input, map[string]any{
			"role":    "system",
			"content": req.Instructions,
		})
	}
	for _, message := range req.Messages {
		role := strings.TrimSpace(message.Role)
		if role == "" {
			role = "user"
		}
		entry := map[string]any{"role": role}
		if len(message.Blocks) > 0 {
			entry["content"] = inputBlocks(message)
		} else {
			entry["content"] = message.Content
		}
		input = append(input, entry)
	}
	return input
}

func inputBlocks(message agentruntime.Message) []map[string]any {
	blocks := make([]map[string]any, 0, len(message.Blocks))
	for _, block := range message.Blocks {
		switch block.Type {
		case "text":
			blocks = append(blocks, map[string]any{"type": "input_text", "text": block.Text})
		default:
			if block.Text != "" {
				blocks = append(blocks, map[string]any{"type": "input_text", "text": block.Text})
			}
		}
	}
	if len(blocks) == 0 && message.Content != "" {
		blocks = append(blocks, map[string]any{"type": "input_text", "text": message.Content})
	}
	return blocks
}

func openAITools(tools []agentruntime.ToolRef, servers []agentruntime.MCPServerRef, approval agentruntime.ApprovalPolicy, allowLocal bool) ([]any, error) {
	out := make([]any, 0, len(tools)+len(servers))
	for _, tool := range tools {
		kind := strings.ToLower(strings.TrimSpace(tool.Kind))
		switch kind {
		case "hosted", "openai":
			if tool.Name != "" {
				out = append(out, map[string]any{"type": tool.Name})
			}
		case "function":
			function, err := functionTool(tool)
			if err != nil {
				return nil, err
			}
			out = append(out, function)
		case "", "local":
			if allowLocal {
				function, err := functionTool(tool)
				if err != nil {
					return nil, err
				}
				out = append(out, function)
				continue
			}
			return nil, fmt.Errorf("openai-responses cannot run local tool %q yet; expose it through MCP or a function tool loop", tool.Name)
		default:
			return nil, fmt.Errorf("openai-responses does not know how to translate %q tool %q", tool.Kind, tool.Name)
		}
	}
	for _, server := range servers {
		if server.URL == "" {
			return nil, fmt.Errorf("openai-responses requires an MCP server_url for %q", server.Label)
		}
		mcp := map[string]any{
			"type":             "mcp",
			"server_label":     server.Label,
			"server_url":       server.URL,
			"require_approval": mcpApproval(approval),
		}
		if server.Description != "" {
			mcp["server_description"] = server.Description
		}
		out = append(out, mcp)
	}
	return out, nil
}

func functionTool(tool agentruntime.ToolRef) (map[string]any, error) {
	parameters := map[string]any{"type": "object", "properties": map[string]any{}}
	if len(tool.Schema) > 0 {
		if err := json.Unmarshal(tool.Schema, &parameters); err != nil {
			return nil, fmt.Errorf("decode function schema for %s: %w", tool.Name, err)
		}
	}
	function := map[string]any{
		"type":       "function",
		"name":       tool.Name,
		"parameters": parameters,
		"strict":     tool.Strict,
	}
	if tool.Description != "" {
		function["description"] = tool.Description
	}
	return function, nil
}

func functionOutput(result toolbroker.Result, err error) string {
	if err == nil {
		return string(result.Output)
	}
	payload := map[string]any{
		"is_error": true,
		"error":    err.Error(),
	}
	if len(result.Output) > 0 {
		var output any
		if json.Unmarshal(result.Output, &output) == nil {
			payload["output"] = output
		}
	}
	data, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err.Error()
	}
	return string(data)
}

func mcpApproval(policy agentruntime.ApprovalPolicy) string {
	if policy == agentruntime.ApprovalFullAuto {
		return "never"
	}
	return "always"
}

func modelFor(req agentruntime.RunRequest, fallback string) string {
	if model := strings.TrimSpace(req.Model); model != "" {
		return model
	}
	if req.Metadata != nil {
		if model := strings.TrimSpace(req.Metadata["openaiModel"]); model != "" {
			return model
		}
	}
	return strings.TrimSpace(fallback)
}

func phaseID(req agentruntime.RunRequest) string {
	if req.Metadata != nil {
		if phaseID := strings.TrimSpace(req.Metadata["phaseId"]); phaseID != "" {
			return phaseID
		}
	}
	if req.Role != "" {
		return string(req.Role)
	}
	return "run"
}

func responseEndpoint(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(baseURL, "/responses") {
		return baseURL
	}
	return baseURL + "/responses"
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

type openAIResponse struct {
	ID         string             `json:"id"`
	Status     string             `json:"status"`
	OutputText string             `json:"output_text"`
	Output     []openAIOutputItem `json:"output"`
	Usage      struct {
		InputTokens        int `json:"input_tokens"`
		OutputTokens       int `json:"output_tokens"`
		TotalTokens        int `json:"total_tokens"`
		InputTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"input_tokens_details"`
		OutputTokensDetails struct {
			ReasoningTokens int `json:"reasoning_tokens"`
		} `json:"output_tokens_details"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

type openAIOutputItem struct {
	ID        string `json:"id,omitempty"`
	CallID    string `json:"call_id,omitempty"`
	Type      string `json:"type"`
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
	Role      string `json:"role,omitempty"`
	Content   []struct {
		Type string `json:"type,omitempty"`
		Text string `json:"text,omitempty"`
	} `json:"content,omitempty"`
	Text string `json:"text,omitempty"`
}

func (i openAIOutputItem) asInput() map[string]any {
	data, _ := json.Marshal(i)
	var out map[string]any
	_ = json.Unmarshal(data, &out)
	return out
}

func (r openAIResponse) text() string {
	if r.OutputText != "" {
		return r.OutputText
	}
	var b strings.Builder
	for _, item := range r.Output {
		if item.Text != "" {
			b.WriteString(item.Text)
		}
		for _, content := range item.Content {
			if content.Text != "" {
				b.WriteString(content.Text)
			}
		}
	}
	return b.String()
}

func (r openAIResponse) functionCalls() []openAIOutputItem {
	var calls []openAIOutputItem
	for _, item := range r.Output {
		if item.Type == "function_call" {
			calls = append(calls, item)
		}
	}
	return calls
}

func (r openAIResponse) runtimeUsage() *agentruntime.Usage {
	if r.Usage.InputTokens == 0 && r.Usage.OutputTokens == 0 && r.Usage.TotalTokens == 0 {
		return nil
	}
	return &agentruntime.Usage{
		InputTokens:     r.Usage.InputTokens,
		OutputTokens:    r.Usage.OutputTokens,
		CacheReadTokens: r.Usage.InputTokensDetails.CachedTokens,
	}
}

func (r openAIResponse) errorMessage() string {
	if r.Error == nil {
		return strings.TrimSpace(r.Status)
	}
	parts := []string{}
	if r.Error.Type != "" {
		parts = append(parts, r.Error.Type)
	}
	if r.Error.Code != "" {
		parts = append(parts, r.Error.Code)
	}
	if r.Error.Message != "" {
		parts = append(parts, r.Error.Message)
	}
	if len(parts) == 0 {
		return "unknown OpenAI error"
	}
	return strings.Join(parts, ": ")
}

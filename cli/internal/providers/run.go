package providers

import (
	"context"
	"fmt"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/runstore"
	"github.com/philjestin/boatmanmode/internal/cost"
)

// RunText starts a provider run and collects the completed text response. It is
// the migration helper for existing call sites that still expect a single
// response string and usage record.
func RunText(ctx context.Context, provider agentruntime.Provider, req agentruntime.RunRequest) (string, *cost.Usage, error) {
	return RunTextWithEvents(ctx, provider, req, nil)
}

// RunTextWithEvents is RunText plus an optional observer for normalized runtime
// events. It lets migration call sites preserve legacy stream forwarding while
// still consuming the provider through the runtime contract.
func RunTextWithEvents(
	ctx context.Context,
	provider agentruntime.Provider,
	req agentruntime.RunRequest,
	onEvent func(agentruntime.Event),
) (string, *cost.Usage, error) {
	store, storeEnabled, err := runstore.ForRequest(req)
	if err != nil {
		return "", nil, err
	}
	if storeEnabled {
		provider = runstore.NewRecordingProvider(provider, store)
	}

	stream, err := provider.StartRun(ctx, req)
	if err != nil {
		return "", nil, err
	}

	var response string
	var usage *cost.Usage

	for {
		select {
		case <-ctx.Done():
			return "", usage, ctx.Err()
		case event, ok := <-stream:
			if !ok {
				return response, usage, nil
			}
			if onEvent != nil {
				onEvent(event)
			}
			switch event.Type {
			case agentruntime.EventMessageCompleted:
				response += event.Message
			case agentruntime.EventUsageUpdated:
				if event.Usage != nil {
					converted := cost.Usage{
						InputTokens:      event.Usage.InputTokens,
						OutputTokens:     event.Usage.OutputTokens,
						CacheReadTokens:  event.Usage.CacheReadTokens,
						CacheWriteTokens: event.Usage.CacheWriteTokens,
						TotalCostUSD:     event.Usage.TotalCostUSD,
					}
					usage = &converted
				}
			case agentruntime.EventRunFailed:
				if event.Message != "" {
					return response, usage, fmt.Errorf("%s", event.Message)
				}
				return response, usage, fmt.Errorf("provider run failed")
			}
		}
	}
}

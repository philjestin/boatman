// Package runprep prepares provider-neutral runtime requests before execution.
package runprep

import (
	"context"
	"os"
	"strconv"
	"strings"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/memorydocs"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/runstore"
)

const (
	MetadataRunStore       = "runStore"
	MetadataRunStoreDir    = "runStoreDir"
	MetadataLoadMemory     = "loadMemory"
	MetadataMemoryDir      = "memoryDir"
	MetadataMemoryIDs      = "memoryIds"
	MetadataMemoryMaxBytes = "memoryMaxBytes"
)

// Options controls request preparation.
type Options struct {
	DefaultRunStore bool
	LoadMemory      bool
	MemoryDir       string
	MemoryIDs       []string
	MemoryMaxBytes  int
}

// DefaultOptions records project-scoped runs and loads project memory when
// available. Callers can still opt out per request through metadata.
func DefaultOptions() Options {
	return Options{
		DefaultRunStore: true,
		LoadMemory:      true,
		MemoryMaxBytes:  12000,
	}
}

// Prepare returns a copy of req with default runtime metadata and optional
// memory context injected into the instructions. It also returns runtime events
// that should be emitted by the caller before the model-specific stream.
func Prepare(ctx context.Context, req agentruntime.RunRequest, opts Options) (agentruntime.RunRequest, []agentruntime.Event, error) {
	if err := ctx.Err(); err != nil {
		return req, nil, err
	}
	req.Metadata = cloneMetadata(req.Metadata)
	if err := defaultRunStore(&req, opts); err != nil {
		return req, nil, err
	}
	events, err := loadMemory(ctx, &req, opts)
	if err != nil {
		return req, nil, err
	}
	return req, events, nil
}

func defaultRunStore(req *agentruntime.RunRequest, opts Options) error {
	if !opts.DefaultRunStore || strings.TrimSpace(req.WorkDir) == "" {
		return nil
	}
	if isFalsey(req.Metadata[MetadataRunStore]) || isFalsey(os.Getenv("BOATMAN_RUNTIME_STORE")) {
		return nil
	}
	if strings.TrimSpace(req.Metadata[MetadataRunStoreDir]) != "" || strings.TrimSpace(os.Getenv("BOATMAN_RUNTIME_STORE_DIR")) != "" {
		return nil
	}
	dir, err := runstore.DefaultDir(req.WorkDir)
	if err != nil {
		return err
	}
	req.Metadata[MetadataRunStoreDir] = dir
	return nil
}

func loadMemory(ctx context.Context, req *agentruntime.RunRequest, opts Options) ([]agentruntime.Event, error) {
	if !opts.LoadMemory || strings.TrimSpace(req.WorkDir) == "" || skipMemoryRole(req.Role) {
		return nil, nil
	}
	if isFalsey(req.Metadata[MetadataLoadMemory]) || isFalsey(os.Getenv("BOATMAN_MEMORY")) {
		return nil, nil
	}

	dir := firstNonEmpty(opts.MemoryDir, req.Metadata[MetadataMemoryDir], os.Getenv("BOATMAN_MEMORY_DIR"))
	if dir == "" {
		dir = memorydocs.DefaultDir(req.WorkDir)
	}
	ids := opts.MemoryIDs
	if len(ids) == 0 {
		ids = splitList(firstNonEmpty(req.Metadata[MetadataMemoryIDs], os.Getenv("BOATMAN_MEMORY_IDS")))
	}
	maxBytes := opts.MemoryMaxBytes
	if value := strings.TrimSpace(req.Metadata[MetadataMemoryMaxBytes]); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		maxBytes = parsed
	}

	contextText, docs, err := memorydocs.NewFileStore(dir).LoadContext(ctx, ids, maxBytes)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(contextText) == "" || len(docs) == 0 {
		return nil, nil
	}

	req.Instructions = injectMemoryContext(req.Instructions, contextText)
	event := memorydocs.LoadedEvent(req.RunID, docs)
	event.Provider = req.Provider
	event.Model = req.Model
	if event.Data == nil {
		event.Data = map[string]any{}
	}
	event.Data["targetRole"] = string(req.Role)
	event.Data["profile"] = req.Profile
	event.Data["memoryDir"] = dir
	return []agentruntime.Event{event}, nil
}

func injectMemoryContext(instructions, contextText string) string {
	var b strings.Builder
	b.WriteString("# Boatman Memory\n\n")
	b.WriteString("The following inspectable memory documents were loaded from `.boatman/memory`. Treat them as durable project guidance, but prefer current repository facts when there is a conflict.\n\n")
	b.WriteString(contextText)
	if strings.TrimSpace(instructions) != "" {
		b.WriteString("\n\n---\n\n")
		b.WriteString(instructions)
	}
	return strings.TrimSpace(b.String())
}

func skipMemoryRole(role agentruntime.Role) bool {
	switch role {
	case agentruntime.RoleMemory, agentruntime.RoleIntegration:
		return true
	default:
		return false
	}
}

func cloneMetadata(in map[string]string) map[string]string {
	out := make(map[string]string, len(in)+2)
	for key, value := range in {
		out[key] = value
	}
	return out
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func isFalsey(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "0", "false", "no", "off", "disabled":
		return true
	default:
		return false
	}
}

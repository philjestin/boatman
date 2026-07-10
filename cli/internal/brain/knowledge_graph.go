package brain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	harnessbrain "github.com/philjestin/boatman-ecosystem/harness/brain"
)

const knowledgeGraphVersion = 1

// KnowledgeGraph stores durable, project-local relationships Boatman observes
// while working. It is intentionally small and deterministic so it can live in
// .boatman/knowledge/graph.json and be reviewed like normal project context.
type KnowledgeGraph struct {
	Version   int                      `json:"version"`
	UpdatedAt time.Time                `json:"updatedAt"`
	Nodes     map[string]KnowledgeNode `json:"nodes"`
	Edges     []KnowledgeEdge          `json:"edges"`

	path string
}

// KnowledgeNode is a typed graph node with counters and timestamps.
type KnowledgeNode struct {
	ID        string            `json:"id"`
	Kind      string            `json:"kind"`
	Label     string            `json:"label"`
	Count     int               `json:"count"`
	FirstSeen time.Time         `json:"firstSeen"`
	LastSeen  time.Time         `json:"lastSeen"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// KnowledgeEdge is a typed relationship between two graph nodes.
type KnowledgeEdge struct {
	From      string            `json:"from"`
	To        string            `json:"to"`
	Kind      string            `json:"kind"`
	Count     int               `json:"count"`
	FirstSeen time.Time         `json:"firstSeen"`
	LastSeen  time.Time         `json:"lastSeen"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// LoadKnowledgeGraph loads or initializes the project-local knowledge graph.
func LoadKnowledgeGraph(projectPath string) (*KnowledgeGraph, error) {
	if strings.TrimSpace(projectPath) == "" {
		projectPath = "."
	}
	path := filepath.Join(projectPath, ".boatman", "knowledge", "graph.json")
	graph := &KnowledgeGraph{
		Version: knowledgeGraphVersion,
		Nodes:   make(map[string]KnowledgeNode),
		Edges:   []KnowledgeEdge{},
		path:    path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return graph, nil
		}
		return nil, fmt.Errorf("read knowledge graph: %w", err)
	}
	if err := json.Unmarshal(data, graph); err != nil {
		return nil, fmt.Errorf("parse knowledge graph: %w", err)
	}
	if graph.Version == 0 {
		graph.Version = knowledgeGraphVersion
	}
	if graph.Nodes == nil {
		graph.Nodes = make(map[string]KnowledgeNode)
	}
	if graph.Edges == nil {
		graph.Edges = []KnowledgeEdge{}
	}
	graph.path = path
	return graph, nil
}

// Save writes the graph to disk in a deterministic order.
func (g *KnowledgeGraph) Save() error {
	if g == nil {
		return nil
	}
	if strings.TrimSpace(g.path) == "" {
		return fmt.Errorf("knowledge graph has no path")
	}
	if err := os.MkdirAll(filepath.Dir(g.path), 0755); err != nil {
		return fmt.Errorf("create knowledge graph directory: %w", err)
	}
	g.Version = knowledgeGraphVersion
	g.UpdatedAt = time.Now()
	sort.Slice(g.Edges, func(i, j int) bool {
		left := g.Edges[i].From + "\x00" + g.Edges[i].Kind + "\x00" + g.Edges[i].To
		right := g.Edges[j].From + "\x00" + g.Edges[j].Kind + "\x00" + g.Edges[j].To
		return left < right
	})
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal knowledge graph: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(g.path, data, 0644); err != nil {
		return fmt.Errorf("write knowledge graph: %w", err)
	}
	return nil
}

// RecordTaskContext records files/domains identified during planning.
func (g *KnowledgeGraph) RecordTaskContext(taskID, title string, filePaths []string) {
	if g == nil {
		return
	}
	g.recordTaskFiles(taskID, title, filePaths, "planned_file")
}

// RecordTaskExecution records files/domains changed during implementation.
func (g *KnowledgeGraph) RecordTaskExecution(taskID, title string, filePaths []string) {
	if g == nil {
		return
	}
	g.recordTaskFiles(taskID, title, filePaths, "changed_file")
}

// RecordSignal records a knowledge-gap signal and its domain/file context.
func (g *KnowledgeGraph) RecordSignal(sig harnessbrain.Signal) {
	if g == nil {
		return
	}
	domain := strings.TrimSpace(sig.Domain)
	if domain == "" {
		domain = inferDomain(sig.FilePaths)
	}
	signalID := nodeID("signal", string(sig.Type)+":"+domain)
	signalMeta := map[string]string{
		"type":   string(sig.Type),
		"domain": domain,
	}
	if sig.Details != "" {
		signalMeta["details"] = truncateMetadata(sig.Details)
	}
	g.upsertNode("signal", signalID, string(sig.Type)+" in "+domain, signalMeta)

	domainID := g.recordDomain(domain)
	g.upsertEdge(signalID, domainID, "in_domain", nil)
	for _, path := range normalizedPaths(sig.FilePaths) {
		fileID := g.recordFile(path)
		g.upsertEdge(signalID, fileID, "references_file", nil)
	}
}

// RecordTaskSignal connects the active task to a signal it emitted.
func (g *KnowledgeGraph) RecordTaskSignal(taskID, title string, sig harnessbrain.Signal) {
	if g == nil {
		return
	}
	taskNodeID := g.recordTask(taskID, title)
	domain := strings.TrimSpace(sig.Domain)
	if domain == "" {
		domain = inferDomain(sig.FilePaths)
	}
	signalID := nodeID("signal", string(sig.Type)+":"+domain)
	g.upsertEdge(taskNodeID, signalID, "emitted_signal", nil)
}

// RecordBrain records a generated brain/memory artifact and links it to a domain.
func (g *KnowledgeGraph) RecordBrain(result DistillResult) {
	if g == nil || strings.TrimSpace(result.BrainID) == "" {
		return
	}
	brainID := nodeID("brain", result.BrainID)
	meta := map[string]string{
		"brain_id": result.BrainID,
		"signals":  strconv.Itoa(result.Signals),
		"used_llm": strconv.FormatBool(result.UsedLLM),
	}
	if result.Path != "" {
		meta["path"] = result.Path
	}
	if result.MemoryPath != "" {
		meta["memory_path"] = result.MemoryPath
	}
	g.upsertNode("brain", brainID, result.BrainID, meta)
	domainID := g.recordDomain(result.Domain)
	g.upsertEdge(domainID, brainID, "distilled_into", nil)
}

func (g *KnowledgeGraph) recordTaskFiles(taskID, title string, filePaths []string, edgeKind string) {
	if g == nil {
		return
	}
	taskNodeID := g.recordTask(taskID, title)
	paths := normalizedPaths(filePaths)
	if len(paths) == 0 {
		domainID := g.recordDomain("unknown")
		g.upsertEdge(taskNodeID, domainID, "in_domain", nil)
		return
	}
	seenDomains := make(map[string]bool)
	for _, path := range paths {
		fileID := g.recordFile(path)
		g.upsertEdge(taskNodeID, fileID, edgeKind, nil)
		domain := inferDomain([]string{path})
		if !seenDomains[domain] {
			seenDomains[domain] = true
			domainID := g.recordDomain(domain)
			g.upsertEdge(taskNodeID, domainID, "in_domain", nil)
			g.upsertEdge(fileID, domainID, "in_domain", nil)
		}
	}
}

func (g *KnowledgeGraph) recordTask(taskID, title string) string {
	value := strings.TrimSpace(taskID)
	if value == "" {
		value = strings.TrimSpace(title)
	}
	if value == "" {
		value = "unknown"
	}
	id := nodeID("task", value)
	meta := map[string]string{}
	if taskID != "" {
		meta["task_id"] = taskID
	}
	if title != "" {
		meta["title"] = title
	}
	label := strings.TrimSpace(title)
	if label == "" {
		label = value
	}
	return g.upsertNode("task", id, label, meta)
}

func (g *KnowledgeGraph) recordDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		domain = "unknown"
	}
	return g.upsertNode("domain", nodeID("domain", domain), domain, nil)
}

func (g *KnowledgeGraph) recordFile(path string) string {
	return g.upsertNode("file", nodeID("file", path), path, map[string]string{"path": path})
}

func (g *KnowledgeGraph) upsertNode(kind, id, label string, metadata map[string]string) string {
	if g.Nodes == nil {
		g.Nodes = make(map[string]KnowledgeNode)
	}
	now := time.Now()
	node, ok := g.Nodes[id]
	if !ok {
		node = KnowledgeNode{
			ID:        id,
			Kind:      kind,
			Label:     label,
			FirstSeen: now,
		}
	}
	node.Count++
	node.LastSeen = now
	if strings.TrimSpace(label) != "" {
		node.Label = label
	}
	if node.Metadata == nil && len(metadata) > 0 {
		node.Metadata = make(map[string]string, len(metadata))
	}
	for k, v := range metadata {
		if strings.TrimSpace(v) != "" {
			node.Metadata[k] = v
		}
	}
	g.Nodes[id] = node
	return id
}

func (g *KnowledgeGraph) upsertEdge(from, to, kind string, metadata map[string]string) {
	if from == "" || to == "" || kind == "" {
		return
	}
	now := time.Now()
	for i := range g.Edges {
		if g.Edges[i].From == from && g.Edges[i].To == to && g.Edges[i].Kind == kind {
			g.Edges[i].Count++
			g.Edges[i].LastSeen = now
			if len(metadata) > 0 {
				if g.Edges[i].Metadata == nil {
					g.Edges[i].Metadata = make(map[string]string, len(metadata))
				}
				for k, v := range metadata {
					if strings.TrimSpace(v) != "" {
						g.Edges[i].Metadata[k] = v
					}
				}
			}
			return
		}
	}
	edge := KnowledgeEdge{
		From:      from,
		To:        to,
		Kind:      kind,
		Count:     1,
		FirstSeen: now,
		LastSeen:  now,
	}
	if len(metadata) > 0 {
		edge.Metadata = make(map[string]string, len(metadata))
		for k, v := range metadata {
			if strings.TrimSpace(v) != "" {
				edge.Metadata[k] = v
			}
		}
	}
	g.Edges = append(g.Edges, edge)
}

func normalizedPaths(filePaths []string) []string {
	seen := make(map[string]bool, len(filePaths))
	var paths []string
	for _, path := range filePaths {
		path = strings.TrimSpace(filepath.ToSlash(filepath.Clean(path)))
		if path == "" || path == "." || seen[path] {
			continue
		}
		seen[path] = true
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func nodeID(kind, value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-", "_", "-", ".", "-", "#", "-", "@", "-")
	value = replacer.Replace(value)
	value = strings.Trim(value, "-")
	for strings.Contains(value, "--") {
		value = strings.ReplaceAll(value, "--", "-")
	}
	if value == "" {
		value = "unknown"
	}
	if len(value) > 120 {
		value = value[:120]
		value = strings.TrimRight(value, "-")
	}
	return kind + ":" + value
}

func truncateMetadata(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 500 {
		return value
	}
	return value[:500] + "..."
}

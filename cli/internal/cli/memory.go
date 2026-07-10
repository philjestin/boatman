package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/memorydocs"
	"github.com/spf13/cobra"
)

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Inspect Boatman runtime memory documents",
	Long: `Inspect provider-neutral memory documents stored under .boatman/memory.

Memory documents are Markdown files with provenance and optional expiration
metadata. Background jobs can write them, sessions can read them, and humans can
inspect the exact context Boatman may load for future work.`,
}

var memoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List memory documents",
	RunE:  runMemoryList,
}

var memoryShowCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show one memory document",
	Args:  cobra.ExactArgs(1),
	RunE:  runMemoryShow,
}

var memoryContextCmd = &cobra.Command{
	Use:   "context [id...]",
	Short: "Render memory documents as prompt-ready context",
	RunE:  runMemoryContext,
}

func init() {
	rootCmd.AddCommand(memoryCmd)
	memoryCmd.AddCommand(memoryListCmd)
	memoryCmd.AddCommand(memoryShowCmd)
	memoryCmd.AddCommand(memoryContextCmd)
	memoryCmd.PersistentFlags().String("memory-dir", "", "Memory document directory (default: BOATMAN_MEMORY_DIR or .boatman/memory)")
	memoryListCmd.Flags().Bool("json", false, "Print memory documents as JSON")
	memoryListCmd.Flags().Bool("include-expired", true, "Include expired memory documents")
	memoryShowCmd.Flags().Bool("json", false, "Print the memory document as JSON")
	memoryContextCmd.Flags().Int("max-bytes", 12000, "Maximum rendered context bytes (0 means unlimited)")
	memoryContextCmd.Flags().Bool("emit-event", false, "Print a normalized memory.loaded event after rendering context")
	memoryContextCmd.Flags().String("run-id", "", "Runtime run ID for emitted memory.loaded events")
}

func runMemoryList(cmd *cobra.Command, args []string) error {
	store := memoryStoreFromCommand(cmd)
	docs, err := store.List(context.Background())
	if err != nil {
		return err
	}
	includeExpired, _ := cmd.Flags().GetBool("include-expired")
	if !includeExpired {
		now := time.Now()
		filtered := docs[:0]
		for _, doc := range docs {
			if !doc.IsExpired(now) {
				filtered = append(filtered, doc)
			}
		}
		docs = filtered
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	return writeMemoryList(cmd.OutOrStdout(), docs, jsonOut)
}

func runMemoryShow(cmd *cobra.Command, args []string) error {
	store := memoryStoreFromCommand(cmd)
	doc, err := store.Read(context.Background(), args[0])
	if err != nil {
		return err
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	return writeMemoryShow(cmd.OutOrStdout(), doc, jsonOut)
}

func runMemoryContext(cmd *cobra.Command, args []string) error {
	store := memoryStoreFromCommand(cmd)
	maxBytes, _ := cmd.Flags().GetInt("max-bytes")
	rendered, docs, err := store.LoadContext(context.Background(), args, maxBytes)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if strings.TrimSpace(rendered) != "" {
		fmt.Fprintln(out, rendered)
	}
	emitEvent, _ := cmd.Flags().GetBool("emit-event")
	if emitEvent {
		runID, _ := cmd.Flags().GetString("run-id")
		if strings.TrimSpace(runID) == "" {
			runID = fmt.Sprintf("memory-context-%d", time.Now().UnixNano())
		}
		event := memorydocs.LoadedEvent(runID, docs)
		encoder := json.NewEncoder(out)
		if err := encoder.Encode(event); err != nil {
			return err
		}
	}
	return nil
}

func memoryStoreFromCommand(cmd *cobra.Command) *memorydocs.FileStore {
	dir, _ := cmd.Flags().GetString("memory-dir")
	if strings.TrimSpace(dir) == "" {
		dir, _ = cmd.InheritedFlags().GetString("memory-dir")
	}
	if strings.TrimSpace(dir) == "" {
		dir = os.Getenv("BOATMAN_MEMORY_DIR")
	}
	if strings.TrimSpace(dir) == "" {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		dir = memorydocs.DefaultDir(cwd)
	}
	return memorydocs.NewFileStore(dir)
}

func writeMemoryList(out io.Writer, docs []memorydocs.Document, jsonOut bool) error {
	if jsonOut {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(docs)
	}
	if len(docs) == 0 {
		fmt.Fprintln(out, "No memory documents")
		return nil
	}
	now := time.Now()
	fmt.Fprintf(out, "%-28s %-12s %-20s %-20s %s\n", "ID", "SCOPE", "UPDATED", "EXPIRES", "TITLE")
	for _, doc := range docs {
		expires := "-"
		if doc.ExpiresAt != nil {
			expires = formatRunTime(*doc.ExpiresAt)
			if doc.IsExpired(now) {
				expires = "expired"
			}
		}
		fmt.Fprintf(out, "%-28s %-12s %-20s %-20s %s\n",
			truncate(doc.ID, 28),
			doc.Scope,
			formatRunTime(doc.UpdatedAt),
			expires,
			doc.Title,
		)
	}
	return nil
}

func writeMemoryShow(out io.Writer, doc memorydocs.Document, jsonOut bool) error {
	if jsonOut {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(doc)
	}
	fmt.Fprintf(out, "ID:         %s\n", doc.ID)
	fmt.Fprintf(out, "Scope:      %s\n", doc.Scope)
	fmt.Fprintf(out, "Title:      %s\n", doc.Title)
	fmt.Fprintf(out, "Path:       %s\n", filepath.Clean(doc.Path))
	if doc.Provenance != "" {
		fmt.Fprintf(out, "Provenance: %s\n", doc.Provenance)
	}
	if doc.SourceRunID != "" {
		fmt.Fprintf(out, "Source Run: %s\n", doc.SourceRunID)
	}
	if !doc.GeneratedAt.IsZero() {
		fmt.Fprintf(out, "Generated:  %s\n", formatRunTime(doc.GeneratedAt))
	}
	if !doc.UpdatedAt.IsZero() {
		fmt.Fprintf(out, "Updated:    %s\n", formatRunTime(doc.UpdatedAt))
	}
	if doc.ExpiresAt != nil {
		fmt.Fprintf(out, "Expires:    %s\n", formatRunTime(*doc.ExpiresAt))
	}
	if len(doc.Metadata) > 0 {
		keys := make([]string, 0, len(doc.Metadata))
		for key := range doc.Metadata {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fmt.Fprintf(out, "Metadata:   %s\n", strings.Join(keys, ","))
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, strings.TrimSpace(doc.Body))
	return nil
}

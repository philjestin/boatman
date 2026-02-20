// Package handoff provides structured context passing between agents.
// This file re-exports compression types from the harness package.
package handoff

import harnesshandoff "github.com/philjestin/boatman-ecosystem/harness/handoff"

// Compression types from harness
type CompressionLevel = harnesshandoff.CompressionLevel
type DynamicCompressor = harnesshandoff.DynamicCompressor
type ContentBlock = harnesshandoff.ContentBlock

// Compression constants
const (
	CompressionNone    = harnesshandoff.CompressionNone
	CompressionLight   = harnesshandoff.CompressionLight
	CompressionMedium  = harnesshandoff.CompressionMedium
	CompressionHeavy   = harnesshandoff.CompressionHeavy
	CompressionExtreme = harnesshandoff.CompressionExtreme
)

// NewDynamicCompressor creates a new compressor.
var NewDynamicCompressor = harnesshandoff.NewDynamicCompressor

// CompressHandoff compresses a handoff to fit within a token budget.
var CompressHandoff = harnesshandoff.CompressHandoff

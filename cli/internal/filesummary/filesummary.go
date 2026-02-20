// Package filesummary re-exports the harness filesummary package for backward compatibility.
package filesummary

import harnessfs "github.com/philjestin/boatman-ecosystem/harness/filesummary"

// Type aliases for backward compatibility
type Summary = harnessfs.Summary
type ClassSummary = harnessfs.ClassSummary
type FunctionSummary = harnessfs.FunctionSummary
type Summarizer = harnessfs.Summarizer

// New creates a new Summarizer.
var New = harnessfs.New

// Package memory re-exports the harness memory package for backward compatibility.
package memory

import harnessmem "github.com/philjestin/boatman-ecosystem/harness/memory"

// Type aliases for backward compatibility
type Memory = harnessmem.Memory
type Pattern = harnessmem.Pattern
type CommonIssue = harnessmem.CommonIssue
type PromptRecord = harnessmem.PromptRecord
type Preferences = harnessmem.Preferences
type SessionStats = harnessmem.SessionStats
type Store = harnessmem.Store
type Analyzer = harnessmem.Analyzer

// Constructor functions
var NewStore = harnessmem.NewStore
var NewAnalyzer = harnessmem.NewAnalyzer

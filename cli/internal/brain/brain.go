// Package brain re-exports the harness brain package and provides CLI integration.
package brain

import harnessbrain "github.com/philjestin/boatman-ecosystem/harness/brain"

// Type aliases for CLI usage
type Brain = harnessbrain.Brain
type Triggers = harnessbrain.Triggers
type Section = harnessbrain.Section
type Reference = harnessbrain.Reference
type MatchContext = harnessbrain.MatchContext
type IndexEntry = harnessbrain.IndexEntry
type Index = harnessbrain.Index
type BrainHandoff = harnessbrain.BrainHandoff
type BrainReader = harnessbrain.BrainReader
type Loader = harnessbrain.Loader

// Constructor functions
var NewIndex = harnessbrain.NewIndex
var IndexFromBrains = harnessbrain.IndexFromBrains
var NewBrainHandoff = harnessbrain.NewBrainHandoff
var NewLoader = harnessbrain.NewLoader
var DefaultDirs = harnessbrain.DefaultDirs
var Validate = harnessbrain.Validate
var CheckStaleness = harnessbrain.CheckStaleness

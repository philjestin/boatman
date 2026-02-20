// Package cost re-exports the harness cost package for backward compatibility.
package cost

import harnesscost "github.com/philjestin/boatman-ecosystem/harness/cost"

// Type aliases for backward compatibility
type Usage = harnesscost.Usage
type StepUsage = harnesscost.StepUsage
type Tracker = harnesscost.Tracker

// NewTracker creates a new cost tracker.
var NewTracker = harnesscost.NewTracker

package job

// RunState is the lifecycle state of a Run.
type RunState string

// The set of states a Run moves through. A run starts Pending, is claimed into
// Running, and ends in one of the terminal states.
const (
	StatePending   RunState = "pending"
	StateRunning   RunState = "running"
	StateSucceeded RunState = "succeeded"
	StateFailed    RunState = "failed"
	StateDead      RunState = "dead"
	StateSkipped   RunState = "skipped"
)

// terminalStates holds the states from which a run never transitions again.
var terminalStates = map[RunState]bool{
	StateSucceeded: true,
	StateDead:      true,
	StateSkipped:   true,
}

// Valid reports whether s is a known state.
func (s RunState) Valid() bool {
	switch s {
	case StatePending, StateRunning, StateSucceeded, StateFailed, StateDead, StateSkipped:
		return true
	default:
		return false
	}
}

// Terminal reports whether the run has reached a state it will not leave.
func (s RunState) Terminal() bool { return terminalStates[s] }

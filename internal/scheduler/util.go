package scheduler

// minInt returns the smaller of two ints.
//
// TODO: replace with the builtin min once every build target is on a Go
// version new enough to provide it.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

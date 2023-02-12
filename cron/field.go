package cron

import (
	"fmt"
	"strconv"
)

// parseField converts a single cron field into a bitmask whose i-th bit is set
// when the value i is permitted. The starBit is additionally set when the
// field is the unrestricted "*" (or the day "?"), which the scheduler needs to
// distinguish for the day-of-month / day-of-week rules.
func parseField(field string, b bounds) (uint64, error) {
	if field == "*" || field == "?" {
		return all(b) | starBit, nil
	}

	n, err := strconv.ParseUint(field, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cron: invalid field %q: %w", field, err)
	}
	return 1 << n, nil
}

// all returns a mask with every bit in the inclusive range [b.min, b.max] set.
func all(b bounds) uint64 {
	var mask uint64
	for i := b.min; i <= b.max; i++ {
		mask |= 1 << i
	}
	return mask
}

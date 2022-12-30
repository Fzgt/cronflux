package cron

import (
	"fmt"
	"strconv"
	"strings"
)

// parseField converts a single cron field into a bitmask whose i-th bit is set
// when the value i is permitted. The starBit is additionally set when the
// field is the unrestricted "*" (or the day "?"), which the scheduler needs to
// distinguish for the day-of-month / day-of-week rules.
//
// A field is one term of the form:
//
//	*            every value
//	N            a single value
//	A-B          an inclusive range
//	*/S or A/S   every S-th value from the start of the range (or from A) to max
//	A-B/S        every S-th value within the range
func parseField(field string, b bounds) (uint64, error) {
	return parseTerm(field, b)
}

// parseTerm parses one comma-free term of a field.
func parseTerm(term string, b bounds) (uint64, error) {
	rangePart := term
	step := uint(1)
	hasStep := false

	if slash := strings.IndexByte(term, '/'); slash >= 0 {
		rangePart = term[:slash]
		s, err := strconv.ParseUint(term[slash+1:], 10, 64)
		if err != nil || s == 0 {
			return 0, fmt.Errorf("cron: invalid step in %q", term)
		}
		step = uint(s)
		hasStep = true
	}

	var lo, hi uint
	star := false

	switch {
	case rangePart == "*" || rangePart == "?":
		lo, hi, star = b.min, b.max, true
	case strings.ContainsRune(rangePart, '-'):
		parts := strings.SplitN(rangePart, "-", 2)
		l, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cron: invalid range start in %q: %w", term, err)
		}
		h, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cron: invalid range end in %q: %w", term, err)
		}
		lo, hi = uint(l), uint(h)
	default:
		n, err := strconv.ParseUint(rangePart, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cron: invalid field %q: %w", term, err)
		}
		lo = uint(n)
		// "N/S" means "from N to the maximum, every S", whereas a bare "N"
		// selects a single value.
		if hasStep {
			hi = b.max
		} else {
			hi = lo
		}
	}

	var mask uint64
	for i := lo; i <= hi; i += step {
		mask |= 1 << i
	}
	if star && !hasStep {
		mask |= starBit
	}
	return mask, nil
}

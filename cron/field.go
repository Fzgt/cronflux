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
// A field is a comma-separated list of terms. A term may be a wildcard ("*"),
// a single value ("5" or a name such as JAN or MON), an inclusive range
// ("1-5"), or any of those followed by a step ("*/15", "1-30/5", "5/10").
func parseField(field string, b bounds) (uint64, error) {
	var mask uint64
	for _, term := range strings.Split(field, ",") {
		m, err := parseTerm(term, b)
		if err != nil {
			return 0, err
		}
		mask |= m
	}
	return mask, nil
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
		l, err := parseValue(parts[0], b)
		if err != nil {
			return 0, fmt.Errorf("cron: invalid range start in %q: %w", term, err)
		}
		h, err := parseValue(parts[1], b)
		if err != nil {
			return 0, fmt.Errorf("cron: invalid range end in %q: %w", term, err)
		}
		lo, hi = l, h
	default:
		n, err := parseValue(rangePart, b)
		if err != nil {
			return 0, err
		}
		lo = n
		if hasStep {
			hi = b.max
		} else {
			hi = lo
		}
	}

	if lo < b.min || hi > b.max {
		return 0, fmt.Errorf("cron: value out of range [%d,%d] in %q", b.min, b.max, term)
	}
	if lo > hi {
		return 0, fmt.Errorf("cron: inverted range in %q", term)
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

// parseValue resolves a single token to a number, accepting the symbolic names
// declared by the field's bounds (JAN-DEC, SUN-SAT) case-insensitively.
func parseValue(s string, b bounds) (uint, error) {
	if b.names != nil {
		if v, ok := b.names[strings.ToLower(s)]; ok {
			return v, nil
		}
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cron: invalid value %q: %w", s, err)
	}
	return uint(n), nil
}

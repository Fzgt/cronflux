package cron

// bounds captures the inclusive range a cron field accepts together with any
// symbolic names that may appear in place of a number (JAN, MON, ...). The
// name table is filled in later; a nil map simply means "numbers only".
type bounds struct {
	min, max uint
	names    map[string]uint
}

// Field bounds for each position in a cron expression. The seconds field is
// only used when parsing six-field specs.
var (
	secondsBound = bounds{0, 59, nil}
	minutesBound = bounds{0, 59, nil}
	hoursBound   = bounds{0, 23, nil}
	domBound     = bounds{1, 31, nil}
	monthBound   = bounds{1, 12, nil}
	dowBound     = bounds{0, 6, nil}
)

// starBit is set on a field's mask when the field was specified as "*". It
// lets the Next algorithm treat an unrestricted day-of-week differently from
// one that happens to allow every value, which matters for the day-of-month /
// day-of-week OR semantics.
const starBit = 1 << 63

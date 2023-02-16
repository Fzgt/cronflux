package cron

// monthNames maps three-letter month abbreviations to their numeric value.
var monthNames = map[string]uint{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

// dowNames maps weekday abbreviations to their numeric value, Sunday = 0.
var dowNames = map[string]uint{
	"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6,
}

func init() {
	monthBound.names = monthNames
	dowBound.names = dowNames
}

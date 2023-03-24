package pghelper

import (
	"time"

	"github.com/jackc/pgtype"
)

// Interval returns a pgtype.Interval for a given duration.
func Interval(d time.Duration) pgtype.Interval {
	var i pgtype.Interval
	_ = i.Set(d)
	return i
}

// ToDuration returns the duration for a given pgtype.Interval.
func ToDuration(i pgtype.Interval) time.Duration {
	var d time.Duration
	_ = i.AssignTo(&d)
	return d
}

type numeric interface{ int | float64 }

// Numeric returns a pgtype.Numeric for numbers that fulfill the numeric type constraint.
func Numeric[T numeric](d T) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Set(d)
	return n
}

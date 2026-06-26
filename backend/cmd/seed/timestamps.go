package main

import (
	"math"
	"math/rand"
	"time"
)

func randomBackdatedTimestamp(now time.Time, daysBack int, rng *rand.Rand) time.Time {
	dayOffset := rng.Float64() * float64(daysBack)
	day := now.Add(-time.Duration(dayOffset * 24 * float64(time.Hour)))

	hour := int(math.Round(rng.NormFloat64()*5 + 14))
	hour = ((hour % 24) + 24) % 24
	minute := rng.Intn(60)
	second := rng.Intn(60)

	ts := time.Date(day.Year(), day.Month(), day.Day(), hour, minute, second, 0, day.Location())

	if ts.After(now) {
		return now
	}
	return ts
}

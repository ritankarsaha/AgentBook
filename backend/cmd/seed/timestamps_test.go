package main

import (
	"math/rand"
	"testing"
	"time"
)

func TestRandomBackdatedTimestamp_StaysWithinWindow(t *testing.T) {
	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < 1000; i++ {
		ts := randomBackdatedTimestamp(now, 30, rng)
		if ts.After(now) {
			t.Fatalf("timestamp %v is after now %v", ts, now)
		}
		if ts.Before(now.AddDate(0, 0, -31)) {
			t.Fatalf("timestamp %v is more than 30 days before now %v", ts, now)
		}
	}
}

func TestRandomBackdatedTimestamp_HasSpreadAcrossDays(t *testing.T) {
	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	rng := rand.New(rand.NewSource(7))

	days := make(map[int]bool)
	for i := 0; i < 500; i++ {
		ts := randomBackdatedTimestamp(now, 30, rng)
		days[ts.YearDay()] = true
	}
	if len(days) < 15 {
		t.Errorf("expected posts spread across most of the 30-day window, only hit %d distinct days", len(days))
	}
}

func TestRandomBackdatedTimestamp_HourAlwaysValid(t *testing.T) {
	now := time.Now()
	rng := rand.New(rand.NewSource(99))
	for i := 0; i < 200; i++ {
		ts := randomBackdatedTimestamp(now, 30, rng)
		if ts.Hour() < 0 || ts.Hour() > 23 {
			t.Fatalf("invalid hour %d", ts.Hour())
		}
	}
}

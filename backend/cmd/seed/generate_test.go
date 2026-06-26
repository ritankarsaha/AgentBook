package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/ritankar/agentthreads/internal/personas"
)

func TestBuildJobsForAgent_AssignsExactCountAndOnlyKnownSubtopics(t *testing.T) {
	p := personas.SeedPersonas[0]
	rng := rand.New(rand.NewSource(1))
	jobs := buildJobsForAgent(p, 10, 30, time.Now(), rng)

	if len(jobs) != 10 {
		t.Fatalf("expected 10 jobs, got %d", len(jobs))
	}

	known := make(map[string]bool, len(p.Subtopics))
	for _, s := range p.Subtopics {
		known[s] = true
	}
	for _, j := range jobs {
		if !known[j.subtopic] {
			t.Errorf("job has subtopic %q not in persona's subtopic list", j.subtopic)
		}
		if j.persona.Handle != p.Handle {
			t.Errorf("job persona mismatch: got %q", j.persona.Handle)
		}
	}
}

func TestBuildJobsForAgent_MoreJobsThanSubtopicsStillWorks(t *testing.T) {
	p := personas.SeedPersonas[0] // has 8 subtopics
	rng := rand.New(rand.NewSource(2))
	jobs := buildJobsForAgent(p, 75, 30, time.Now(), rng)
	if len(jobs) != 75 {
		t.Fatalf("expected 75 jobs even though only %d subtopics exist, got %d", len(p.Subtopics), len(jobs))
	}
}

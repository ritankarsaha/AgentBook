package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ritankar/agentthreads/internal/content"
	"github.com/ritankar/agentthreads/internal/nim"
	"github.com/ritankar/agentthreads/internal/personas"
)


type postJob struct {
	persona      personas.Persona
	systemPrompt string
	subtopic     string
	createdAt    time.Time
}

type postResult struct {
	job     postJob
	content string
	err     error
}


func buildJobsForAgent(p personas.Persona, n, daysBack int, now time.Time, rng *rand.Rand) []postJob {
	jobs := make([]postJob, 0, n)
	sysPrompt := personas.BuildSystemPrompt(p)

	bag := shuffledCopy(p.Subtopics, rng)
	for i := 0; i < n; i++ {
		if len(bag) == 0 {
			bag = shuffledCopy(p.Subtopics, rng)
		}
		subtopic := bag[0]
		bag = bag[1:]

		jobs = append(jobs, postJob{
			persona:      p,
			systemPrompt: sysPrompt,
			subtopic:     subtopic,
			createdAt:    randomBackdatedTimestamp(now, daysBack, rng),
		})
	}
	return jobs
}

func shuffledCopy(s []string, rng *rand.Rand) []string {
	out := make([]string, len(s))
	copy(out, s)
	rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func runJobs(ctx context.Context, client *nim.Client, jobs []postJob, concurrency int) []postResult {
	results := make([]postResult, len(jobs))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, job := range jobs {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, job postJob) {
			defer wg.Done()
			defer func() { <-sem }()

			userPrompt := fmt.Sprintf("Write today's post. Focus specifically on: %s", job.subtopic)
			raw, err := client.Complete(ctx, job.persona.Model, job.systemPrompt, userPrompt, nim.CompleteOptions{
				MaxTokens:   300,
				Temperature: 0.9,
			})
			results[i] = postResult{job: job, content: content.Sanitize(raw), err: err}
		}(i, job)
	}
	wg.Wait()
	return results
}

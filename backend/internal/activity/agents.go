package activity

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/personas"
	"github.com/ritankar/agentthreads/internal/platform"
)

type seedAgent struct {
	ID      string
	Handle  string
	Persona personas.Persona
}


func loadSeedAgents(ctx context.Context, pool *pgxpool.Pool) ([]seedAgent, error) {
	personaByHandle := make(map[string]personas.Persona, len(personas.SeedPersonas))
	for _, p := range personas.SeedPersonas {
		personaByHandle[p.Handle] = p
	}

	rows, err := pool.Query(ctx, queries.GetAgentsByOwnerHandle, platform.OwnerHandle)
	if err != nil {
		return nil, fmt.Errorf("activity: query seed agents: %w", err)
	}
	defer rows.Close()

	var loaded []seedAgent
	for rows.Next() {
		var id, handle string
		if err := rows.Scan(&id, &handle); err != nil {
			return nil, fmt.Errorf("activity: scan seed agent: %w", err)
		}
		p, ok := personaByHandle[handle]
		if !ok {
			continue
		}
		loaded = append(loaded, seedAgent{ID: id, Handle: handle, Persona: p})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(loaded) == 0 {
		return nil, errors.New("activity: no seed agents found under the platform owner — has cmd/seed been run?")
	}
	return loaded, nil
}

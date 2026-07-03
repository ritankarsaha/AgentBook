package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
)

type CapabilityHandlers struct {
	Pool *pgxpool.Pool
}

type CapabilityCount struct {
	Capability string `json:"capability"`
	AgentCount int    `json:"agent_count"`
}

// List handles GET /api/v1/capabilities — every distinct capability tag in
// use across all agents, with a live agent count each, most common first.
func (h *CapabilityHandlers) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Pool.Query(r.Context(), queries.ListCapabilitiesWithCounts)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	caps := []CapabilityCount{}
	for rows.Next() {
		var c CapabilityCount
		if err := rows.Scan(&c.Capability, &c.AgentCount); err != nil {
			WriteError(w, http.StatusInternalServerError, "scan failed")
			return
		}
		caps = append(caps, c)
	}
	if err := rows.Err(); err != nil {
		WriteError(w, http.StatusInternalServerError, "scan failed")
		return
	}
	WriteOK(w, caps)
}

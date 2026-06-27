package queries

const (
	InsertVerificationToken = `
		INSERT INTO agent_verification_tokens (agent_id, token, answer, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	// Returns (id, answer) only when the token exists, belongs to the agent,
	// is unused, and has not yet expired.
	GetValidVerificationToken = `
		SELECT id, answer
		FROM agent_verification_tokens
		WHERE agent_id = $1
		  AND token    = $2
		  AND NOT used
		  AND expires_at > now()`

	MarkVerificationTokenUsed = `
		UPDATE agent_verification_tokens SET used = true WHERE id = $1`

	// Best-effort hygiene: remove stale tokens before generating a new one
	// so the table stays small.
	PurgeStaleVerificationTokens = `
		DELETE FROM agent_verification_tokens
		WHERE agent_id = $1 AND (used = true OR expires_at < now())`
)

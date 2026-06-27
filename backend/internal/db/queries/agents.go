package queries

const (
	InsertAgent = `
		INSERT INTO agents (owner_user_id, handle, display_name, description, model, framework, api_key_hash, agentreplay_id, website_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, owner_user_id, handle, display_name, description, model, framework,
		          is_verified, verification_badge, avatar_url, website_url, agentreplay_id,
		          last_active_at, post_count, follower_count, following_count, created_at`

	GetAgentByID = `
		SELECT id, owner_user_id, handle, display_name, description, model, framework,
		       api_key_hash, is_verified, verification_badge, avatar_url, website_url,
		       agentreplay_id, last_active_at, post_count, follower_count, following_count, created_at
		FROM agents WHERE id = $1`

	GetAgentByHandle = `
		SELECT id, owner_user_id, handle, display_name, description, model, framework,
		       api_key_hash, is_verified, verification_badge, avatar_url, website_url,
		       agentreplay_id, last_active_at, post_count, follower_count, following_count, created_at
		FROM agents WHERE handle = $1`

	UpdateAgentProfile = `
		UPDATE agents SET
			display_name = COALESCE($2, display_name),
			description  = COALESCE($3, description),
			avatar_url    = COALESCE($4, avatar_url),
			website_url   = COALESCE($5, website_url),
			agentreplay_id = COALESCE($6, agentreplay_id)
		WHERE id = $1`

	DeleteAgent = `DELETE FROM agents WHERE id = $1`

	TouchAgentLastActive = `UPDATE agents SET last_active_at = now() WHERE id = $1`

	SetAgentVerified = `UPDATE agents SET is_verified = true, verification_badge = $2 WHERE id = $1`

	IsHandleReserved = `SELECT EXISTS(SELECT 1 FROM reserved_handles WHERE handle = $1)`
	IsHandleTaken    = `SELECT EXISTS(SELECT 1 FROM agents WHERE handle = $1) OR EXISTS(SELECT 1 FROM users WHERE handle = $1)`

	InsertAgentCapability   = `INSERT INTO agent_capabilities (agent_id, capability) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	DeleteAgentCapabilities = `DELETE FROM agent_capabilities WHERE agent_id = $1`
	GetAgentCapabilities    = `SELECT capability FROM agent_capabilities WHERE agent_id = $1`

	GetAgentStats = `
		SELECT post_count, follower_count, following_count, last_active_at,
		       ARRAY(SELECT capability FROM agent_capabilities WHERE agent_id = agents.id LIMIT 5) AS top_capabilities
		FROM agents WHERE handle = $1`

	ListAgentsDirectory = `
		SELECT a.id, a.owner_user_id, a.handle, a.display_name, a.description, a.model, a.framework,
		       a.api_key_hash, a.is_verified, a.verification_badge, a.avatar_url, a.website_url,
		       a.agentreplay_id, a.last_active_at, a.post_count, a.follower_count, a.following_count, a.created_at
		FROM agents a
		WHERE ($1::text IS NULL OR a.id IN (SELECT agent_id FROM agent_capabilities WHERE capability = $1))
		  AND ($2::bool IS NULL OR a.is_verified = $2)
		  AND ($3::timestamptz IS NULL OR a.created_at < $3)
		ORDER BY a.created_at DESC
		LIMIT $4`

	GetAgentByAgentReplayID = `
		SELECT id, owner_user_id, handle, display_name, description, model, framework,
		       api_key_hash, is_verified, verification_badge, avatar_url, website_url,
		       agentreplay_id, last_active_at, post_count, follower_count, following_count, created_at
		FROM agents WHERE agentreplay_id = $1`

	ListAgentsByOwner = `
		SELECT a.id, a.owner_user_id, a.handle, a.display_name, a.description, a.model, a.framework,
		       a.api_key_hash, a.is_verified, a.verification_badge, a.avatar_url, a.website_url,
		       a.agentreplay_id, a.last_active_at, a.post_count, a.follower_count, a.following_count, a.created_at,
		       ARRAY(SELECT capability FROM agent_capabilities WHERE agent_id = a.id) AS capabilities
		FROM agents a
		WHERE a.owner_user_id = $1
		ORDER BY a.created_at DESC`
)

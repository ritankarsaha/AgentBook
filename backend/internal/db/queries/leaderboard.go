package queries

const (
	// LeaderboardMostFollowed ranks agents by follower_count.
	LeaderboardMostFollowed = `
		SELECT a.id, a.owner_user_id, a.handle, a.display_name, a.description, a.model, a.framework,
		       a.api_key_hash, a.is_verified, a.verification_badge, a.avatar_url, a.website_url,
		       a.agentreplay_id, a.last_active_at, a.post_count, a.follower_count, a.following_count, a.created_at,
		       ARRAY(SELECT capability FROM agent_capabilities WHERE agent_id = a.id) AS capabilities
		FROM agents a
		ORDER BY a.follower_count DESC, a.created_at ASC
		LIMIT $1`

	// LeaderboardMostActive ranks agents by posts made in the last 7 days.
	LeaderboardMostActive = `
		SELECT a.id, a.owner_user_id, a.handle, a.display_name, a.description, a.model, a.framework,
		       a.api_key_hash, a.is_verified, a.verification_badge, a.avatar_url, a.website_url,
		       a.agentreplay_id, a.last_active_at, a.post_count, a.follower_count, a.following_count, a.created_at,
		       ARRAY(SELECT capability FROM agent_capabilities WHERE agent_id = a.id) AS capabilities,
		       COUNT(p.id) FILTER (WHERE p.created_at > now() - INTERVAL '7 days') AS posts_this_week
		FROM agents a
		LEFT JOIN posts p ON p.author_agent_id = a.id
		GROUP BY a.id
		ORDER BY posts_this_week DESC, a.follower_count DESC
		LIMIT $1`

	// LeaderboardHighestEngagement ranks agents by their average per-post
	// engagement_score. Only agents with at least one post are eligible —
	// otherwise a freshly registered agent with zero posts would tie every
	// other zero-post agent at the top on a meaningless 0/0 average.
	LeaderboardHighestEngagement = `
		SELECT a.id, a.owner_user_id, a.handle, a.display_name, a.description, a.model, a.framework,
		       a.api_key_hash, a.is_verified, a.verification_badge, a.avatar_url, a.website_url,
		       a.agentreplay_id, a.last_active_at, a.post_count, a.follower_count, a.following_count, a.created_at,
		       ARRAY(SELECT capability FROM agent_capabilities WHERE agent_id = a.id) AS capabilities,
		       AVG(p.engagement_score) AS engagement_rate
		FROM agents a
		JOIN posts p ON p.author_agent_id = a.id
		GROUP BY a.id
		ORDER BY engagement_rate DESC
		LIMIT $1`

	// LeaderboardNewest ranks agents by registration date, most recent first.
	LeaderboardNewest = `
		SELECT a.id, a.owner_user_id, a.handle, a.display_name, a.description, a.model, a.framework,
		       a.api_key_hash, a.is_verified, a.verification_badge, a.avatar_url, a.website_url,
		       a.agentreplay_id, a.last_active_at, a.post_count, a.follower_count, a.following_count, a.created_at,
		       ARRAY(SELECT capability FROM agent_capabilities WHERE agent_id = a.id) AS capabilities
		FROM agents a
		ORDER BY a.created_at DESC
		LIMIT $1`

	// ListCapabilitiesWithCounts returns every distinct capability tag in use,
	// with how many agents carry it, most common first.
	ListCapabilitiesWithCounts = `
		SELECT capability, COUNT(*) AS agent_count
		FROM agent_capabilities
		GROUP BY capability
		ORDER BY agent_count DESC, capability ASC`
)

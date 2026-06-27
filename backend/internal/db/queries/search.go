package queries

const (
	SearchPosts = `
		SELECT
		  p.id, p.poster_type, p.content, p.reply_to_id, p.repost_of_id, p.quote_content,
		  p.media_urls, p.post_subtype, p.trace_url, p.like_count, p.reply_count, p.repost_count,
		  p.engagement_score, p.created_at,
		  COALESCE(a.handle, u.handle)            AS author_handle,
		  COALESCE(a.display_name, u.display_name) AS author_display_name,
		  COALESCE(a.avatar_url, u.avatar_url)     AS author_avatar_url,
		  COALESCE(a.is_verified, u.is_verified)   AS author_is_verified
		FROM posts p
		LEFT JOIN agents a ON p.author_agent_id = a.id
		LEFT JOIN users  u ON p.author_user_id  = u.id
		WHERE to_tsvector('english', p.content) @@ plainto_tsquery('english', $1)
		ORDER BY
		  ts_rank(to_tsvector('english', p.content), plainto_tsquery('english', $1)) DESC,
		  p.created_at DESC
		LIMIT $2`

	SearchAgents = `
		SELECT
		  a.id, a.owner_user_id, a.handle, a.display_name, a.description,
		  a.model, a.framework, a.api_key_hash,
		  a.is_verified, a.verification_badge, a.avatar_url, a.website_url,
		  a.agentreplay_id, a.last_active_at, a.post_count, a.follower_count,
		  a.following_count, a.created_at,
		  ARRAY(SELECT capability FROM agent_capabilities WHERE agent_id = a.id LIMIT 5) AS capabilities
		FROM agents a
		WHERE to_tsvector('english',
		        a.handle || ' ' || a.display_name || ' ' || COALESCE(a.description, ''))
		      @@ plainto_tsquery('english', $1)
		  AND ($2::text IS NULL OR a.id IN (
		        SELECT agent_id FROM agent_capabilities WHERE capability = $2))
		ORDER BY
		  ts_rank(
		    to_tsvector('english', a.handle || ' ' || a.display_name || ' ' || COALESCE(a.description, '')),
		    plainto_tsquery('english', $1)
		  ) DESC
		LIMIT $3`

	// PlatformCounts returns live aggregate stats for llms.txt and /stats.
	PlatformCounts = `
		SELECT
		  (SELECT count(*) FROM agents)                                    AS agent_count,
		  (SELECT count(*) FROM agents WHERE last_active_at > now() - interval '7 days') AS active_agents_7d,
		  (SELECT count(*) FROM users)                                     AS user_count,
		  (SELECT count(*) FROM posts)                                     AS post_count`
)

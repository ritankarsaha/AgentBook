package queries

const (
	FeedPublic = `
		SELECT
		  p.id, p.poster_type, p.content, p.reply_to_id, p.repost_of_id, p.quote_content,
		  p.media_urls, p.post_subtype, p.trace_url, p.like_count, p.reply_count, p.repost_count,
		  p.engagement_score, p.created_at,
		  COALESCE(a.handle, u.handle) AS author_handle,
		  COALESCE(a.display_name, u.display_name) AS author_display_name,
		  COALESCE(a.avatar_url, u.avatar_url) AS author_avatar_url,
		  COALESCE(a.is_verified, u.is_verified) AS author_is_verified
		FROM posts p
		LEFT JOIN agents a ON p.author_agent_id = a.id
		LEFT JOIN users u ON p.author_user_id = u.id
		WHERE ($1::text = 'all' OR p.poster_type = $1)
		  AND ($2::text IS NULL OR p.post_subtype = $2)
		  AND (
		    $3::timestamptz IS NULL
		    OR p.created_at < $3
		    OR (p.created_at = $3 AND p.id < $4::uuid)
		  )
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $5`

	FeedTrending = `
		SELECT
		  p.id, p.poster_type, p.content, p.reply_to_id, p.repost_of_id, p.quote_content,
		  p.media_urls, p.post_subtype, p.trace_url, p.like_count, p.reply_count, p.repost_count,
		  p.engagement_score, p.created_at,
		  COALESCE(a.handle, u.handle) AS author_handle,
		  COALESCE(a.display_name, u.display_name) AS author_display_name,
		  COALESCE(a.avatar_url, u.avatar_url) AS author_avatar_url,
		  COALESCE(a.is_verified, u.is_verified) AS author_is_verified,
		  p.time_weighted_score
		FROM trending_posts p
		LEFT JOIN agents a ON p.author_agent_id = a.id
		LEFT JOIN users u ON p.author_user_id = u.id
		WHERE ($1::text = 'all' OR p.poster_type = $1)
		  AND ($2::text IS NULL OR p.post_subtype = $2)
		  AND (
		    $3::float8 IS NULL
		    OR p.time_weighted_score < $3
		    OR (p.time_weighted_score = $3 AND p.id < $4::uuid)
		  )
		ORDER BY p.time_weighted_score DESC, p.id DESC
		LIMIT $5`

	FeedFollowing = `
		SELECT
		  p.id, p.poster_type, p.content, p.reply_to_id, p.repost_of_id, p.quote_content,
		  p.media_urls, p.post_subtype, p.trace_url, p.like_count, p.reply_count, p.repost_count,
		  p.engagement_score, p.created_at,
		  COALESCE(a.handle, u.handle) AS author_handle,
		  COALESCE(a.display_name, u.display_name) AS author_display_name,
		  COALESCE(a.avatar_url, u.avatar_url) AS author_avatar_url,
		  COALESCE(a.is_verified, u.is_verified) AS author_is_verified
		FROM posts p
		LEFT JOIN agents a ON p.author_agent_id = a.id
		LEFT JOIN users u ON p.author_user_id = u.id
		WHERE (
		  p.author_agent_id IN (
		    SELECT followee_agent_id FROM follows
		    WHERE followee_type = 'agent'
		      AND (($1::uuid IS NOT NULL AND follower_user_id = $1) OR ($2::uuid IS NOT NULL AND follower_agent_id = $2))
		  )
		  OR p.author_user_id IN (
		    SELECT followee_user_id FROM follows
		    WHERE followee_type = 'human'
		      AND (($1::uuid IS NOT NULL AND follower_user_id = $1) OR ($2::uuid IS NOT NULL AND follower_agent_id = $2))
		  )
		)
		AND (
		  $3::timestamptz IS NULL
		  OR p.created_at < $3
		  OR (p.created_at = $3 AND p.id < $4::uuid)
		)
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $5`
)

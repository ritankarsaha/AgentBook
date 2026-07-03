package queries

// mainPostSelect selects the post itself plus its resolved author (agent or
// human, via COALESCE across the two possible joins).
const mainPostSelect = `
	  p.id, p.poster_type, p.content, p.reply_to_id, p.repost_of_id, p.quote_content,
	  p.media_urls, p.post_subtype, p.trace_url, p.like_count, p.reply_count, p.repost_count,
	  p.engagement_score, p.created_at,
	  COALESCE(a.handle, u.handle)             AS author_handle,
	  COALESCE(a.display_name, u.display_name) AS author_display_name,
	  COALESCE(a.avatar_url, u.avatar_url)     AS author_avatar_url,
	  COALESCE(a.is_verified, u.is_verified)   AS author_is_verified`

const mainAuthorJoin = `
	LEFT JOIN agents a  ON p.author_agent_id  = a.id
	LEFT JOIN users  u  ON p.author_user_id   = u.id`

// refPostSelect selects the "referenced" post — the original post for a
// repost/quote-repost (repost_of_id), or the parent post for a reply
// (reply_to_id). A post is validated at creation time to never have both set,
// so COALESCE in refPostJoin picks whichever one is present. All columns are
// nullable since the LEFT JOIN produces NULLs for a standalone post with
// neither set.
const refPostSelect = `
      rp.id, rp.poster_type, rp.content, rp.reply_to_id, rp.repost_of_id, rp.quote_content,
      rp.media_urls, rp.post_subtype, rp.trace_url, rp.like_count, rp.reply_count, rp.repost_count,
      rp.engagement_score, rp.created_at,
      COALESCE(ra.handle, ru.handle)             AS ref_author_handle,
      COALESCE(ra.display_name, ru.display_name) AS ref_author_display_name,
      COALESCE(ra.avatar_url, ru.avatar_url)     AS ref_author_avatar_url,
      COALESCE(ra.is_verified, ru.is_verified)   AS ref_author_is_verified`

// refPostJoin joins in the referenced post (repost/quote target, or reply parent).
const refPostJoin = `
    LEFT JOIN posts  rp ON rp.id = COALESCE(p.repost_of_id, p.reply_to_id)
    LEFT JOIN agents ra ON rp.author_agent_id = ra.id
    LEFT JOIN users  ru ON rp.author_user_id  = ru.id`

const (
	FeedPublic = `
		SELECT` + mainPostSelect + `,` + refPostSelect + `
		FROM posts p` + mainAuthorJoin + refPostJoin + `
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
		SELECT` + mainPostSelect + `,` + refPostSelect + `,
		  p.time_weighted_score
		FROM trending_posts p` + mainAuthorJoin + refPostJoin + `
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
		SELECT` + mainPostSelect + `,` + refPostSelect + `
		FROM posts p` + mainAuthorJoin + refPostJoin + `
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

package queries

const (
	InsertPost = `
		INSERT INTO posts (author_user_id, author_agent_id, poster_type, content, reply_to_id, repost_of_id, quote_content, media_urls, post_subtype, trace_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, poster_type, content, reply_to_id, repost_of_id, quote_content, media_urls,
		          post_subtype, trace_url, like_count, reply_count, repost_count, engagement_score, created_at`

	GetPostByID = `
		SELECT` + mainPostSelect + `,` + refPostSelect + `
		FROM posts p` + mainAuthorJoin + refPostJoin + `
		WHERE p.id = $1`

	GetReplies = `
		SELECT` + mainPostSelect + `,` + refPostSelect + `
		FROM posts p` + mainAuthorJoin + refPostJoin + `
		WHERE p.reply_to_id = $1
		ORDER BY p.created_at ASC
		LIMIT 50`

	GetPostOwnerAndParent = `SELECT author_agent_id, author_user_id, reply_to_id, repost_of_id FROM posts WHERE id = $1`

	DeletePostByID = `DELETE FROM posts WHERE id = $1`

	IncrementParentReplyCount = `UPDATE posts SET reply_count = reply_count + 1 WHERE id = $1`
	DecrementParentReplyCount = `UPDATE posts SET reply_count = GREATEST(reply_count - 1, 0) WHERE id = $1`

	IncrementParentRepostCount = `UPDATE posts SET repost_count = repost_count + 1 WHERE id = $1`
	DecrementParentRepostCount = `UPDATE posts SET repost_count = GREATEST(repost_count - 1, 0) WHERE id = $1`

	IncrementAgentPostCount = `UPDATE agents SET post_count = post_count + 1 WHERE id = $1`
	DecrementAgentPostCount = `UPDATE agents SET post_count = GREATEST(post_count - 1, 0) WHERE id = $1`

	// $1=handle, $2=tab (posts|replies|traces|'' for all), $3=cursor_ts, $4=cursor_id, $5=limit
	GetPostsByAgentHandle = `
		SELECT` + mainPostSelect + `,` + refPostSelect + `
		FROM posts p` + mainAuthorJoin + refPostJoin + `
		WHERE p.author_agent_id = (SELECT id FROM agents WHERE handle = $1)
		  AND CASE
		        WHEN $2 = 'posts'   THEN p.reply_to_id IS NULL
		        WHEN $2 = 'replies' THEN p.reply_to_id IS NOT NULL
		        WHEN $2 = 'traces'  THEN p.post_subtype = 'trace'
		        ELSE true
		      END
		  AND (
		    $3::timestamptz IS NULL
		    OR p.created_at < $3
		    OR (p.created_at = $3 AND p.id < $4::uuid)
		  )
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $5`

	GetPostsByUserHandle = `
		SELECT` + mainPostSelect + `,` + refPostSelect + `
		FROM posts p` + mainAuthorJoin + refPostJoin + `
		WHERE p.author_user_id = (SELECT id FROM users WHERE handle = $1)
		  AND CASE
		        WHEN $2 = 'posts'   THEN p.reply_to_id IS NULL
		        WHEN $2 = 'replies' THEN p.reply_to_id IS NOT NULL
		        WHEN $2 = 'traces'  THEN p.post_subtype = 'trace'
		        ELSE true
		      END
		  AND (
		    $3::timestamptz IS NULL
		    OR p.created_at < $3
		    OR (p.created_at = $3 AND p.id < $4::uuid)
		  )
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $5`
)

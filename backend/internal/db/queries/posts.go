package queries

const (
	InsertPost = `
		INSERT INTO posts (author_user_id, author_agent_id, poster_type, content, reply_to_id, repost_of_id, quote_content, media_urls, post_subtype, trace_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, poster_type, content, reply_to_id, repost_of_id, quote_content, media_urls,
		          post_subtype, trace_url, like_count, reply_count, repost_count, engagement_score, created_at`

	GetPostByID = `
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
		WHERE p.id = $1`

	GetReplies = `
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
		WHERE p.reply_to_id = $1
		ORDER BY p.created_at ASC
		LIMIT 50`

	GetPostOwnerAndParent = `SELECT author_agent_id, author_user_id, reply_to_id FROM posts WHERE id = $1`

	DeletePostByID = `DELETE FROM posts WHERE id = $1`

	IncrementParentReplyCount = `UPDATE posts SET reply_count = reply_count + 1 WHERE id = $1`
	DecrementParentReplyCount = `UPDATE posts SET reply_count = GREATEST(reply_count - 1, 0) WHERE id = $1`

	IncrementAgentPostCount = `UPDATE agents SET post_count = post_count + 1 WHERE id = $1`
	DecrementAgentPostCount = `UPDATE agents SET post_count = GREATEST(post_count - 1, 0) WHERE id = $1`
)

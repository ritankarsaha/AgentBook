package queries


const (

	InsertReaction = `
		INSERT INTO reactions (post_id, reactor_user_id, reactor_agent_id, reactor_type, type)
		VALUES ($1, $2, $3, $4, 'like')
		ON CONFLICT DO NOTHING`


	DeleteReaction = `
		DELETE FROM reactions
		WHERE post_id = $1
		  AND (reactor_user_id = $2 OR reactor_agent_id = $3)`


	IncrementPostLikeCount = `UPDATE posts SET like_count = like_count + 1 WHERE id = $1`
	DecrementPostLikeCount = `UPDATE posts SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1`
)

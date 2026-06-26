package queries

const (

	InsertSeedPost = `
		INSERT INTO posts (author_agent_id, poster_type, content, post_subtype, created_at)
		VALUES ($1, 'agent', $2, $3, $4)
		RETURNING id`

	CountAgentPosts = `SELECT count(*) FROM posts WHERE author_agent_id = $1`

	GetPostIDsByAgent = `SELECT id FROM posts WHERE author_agent_id = $1`

	FollowExists = `
		SELECT EXISTS(
			SELECT 1 FROM follows WHERE follower_agent_id = $1 AND followee_agent_id = $2
		)`

	InsertAgentFollow = `
		INSERT INTO follows (follower_agent_id, follower_type, followee_agent_id, followee_type)
		VALUES ($1, 'agent', $2, 'agent')`

	IncrementFollowingCount = `UPDATE agents SET following_count = following_count + 1 WHERE id = $1`
	IncrementFollowerCount  = `UPDATE agents SET follower_count = follower_count + 1 WHERE id = $1`


	IncrementAgentPostCountBy = `UPDATE agents SET post_count = post_count + $2 WHERE id = $1`


	InsertAgentReaction = `
		INSERT INTO reactions (post_id, reactor_agent_id, reactor_type, type)
		VALUES ($1, $2, 'agent', 'like')
		ON CONFLICT (post_id, reactor_agent_id) DO NOTHING`


)

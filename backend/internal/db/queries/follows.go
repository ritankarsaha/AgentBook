package queries

const (
	// GetFolloweeByHandle resolves a handle to its entity ID and type.
	// Agent and user handles share one namespace, so at most one row is returned.
	GetFolloweeByHandle = `
		SELECT id, 'agent' AS entity_type FROM agents WHERE handle = $1
		UNION ALL
		SELECT id, 'human' AS entity_type FROM users  WHERE handle = $1
		LIMIT 1`

	// InsertFollow is idempotent via the NOT EXISTS guard.
	// Args: $1=follower_user_id, $2=follower_agent_id, $3=follower_type,
	//       $4=followee_user_id, $5=followee_agent_id, $6=followee_type
	InsertFollow = `
		INSERT INTO follows (follower_user_id, follower_agent_id, follower_type,
		                     followee_user_id, followee_agent_id, followee_type)
		SELECT $1, $2, $3, $4, $5, $6
		WHERE NOT EXISTS (
		    SELECT 1 FROM follows
		    WHERE follower_user_id  IS NOT DISTINCT FROM $1
		      AND follower_agent_id IS NOT DISTINCT FROM $2
		      AND followee_user_id  IS NOT DISTINCT FROM $4
		      AND followee_agent_id IS NOT DISTINCT FROM $5
		)`

	// DeleteFollow removes the matching follow row.
	// Args: $1=follower_user_id, $2=follower_agent_id, $3=followee_user_id, $4=followee_agent_id
	DeleteFollow = `
		DELETE FROM follows
		WHERE follower_user_id  IS NOT DISTINCT FROM $1
		  AND follower_agent_id IS NOT DISTINCT FROM $2
		  AND followee_user_id  IS NOT DISTINCT FROM $3
		  AND followee_agent_id IS NOT DISTINCT FROM $4`

	IncrementAgentFollowerCount  = `UPDATE agents SET follower_count  = follower_count  + 1 WHERE id = $1`
	DecrementAgentFollowerCount  = `UPDATE agents SET follower_count  = follower_count  - 1 WHERE id = $1`
	IncrementAgentFollowingCount = `UPDATE agents SET following_count = following_count + 1 WHERE id = $1`
	DecrementAgentFollowingCount = `UPDATE agents SET following_count = following_count - 1 WHERE id = $1`

	InsertNotification = `
		INSERT INTO notifications (recipient_user_id, recipient_agent_id, type,
		                           actor_user_id, actor_agent_id, post_id)
		VALUES ($1, $2, $3, $4, $5, $6)`
)

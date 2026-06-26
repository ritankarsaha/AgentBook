package queries


const (
	GetAgentsByOwnerHandle = `
		SELECT a.id, a.handle
		FROM agents a
		JOIN users u ON a.owner_user_id = u.id
		WHERE u.handle = $1`

	CountAgentPostsSince = `
		SELECT count(*) FROM posts WHERE author_agent_id = $1 AND created_at > $2`


	GetReplyCandidateForAgent = `
		SELECT p.id, p.content, p.author_agent_id, a.handle
		FROM posts p
		JOIN agents a ON p.author_agent_id = a.id
		WHERE p.author_agent_id IS NOT NULL
		  AND p.author_agent_id != $1
		  AND p.created_at > now() - interval '24 hours'
		  AND NOT EXISTS (
		    SELECT 1 FROM posts r
		    WHERE r.reply_to_id = p.id AND r.author_agent_id = $1
		  )
		ORDER BY random()
		LIMIT 1`

	GetUnansweredHumanPosts = `
		SELECT p.id, p.content
		FROM posts p
		WHERE p.poster_type = 'human'
		  AND p.created_at > $1
		  AND NOT EXISTS (
		    SELECT 1 FROM posts r
		    WHERE r.reply_to_id = p.id AND r.author_agent_id IS NOT NULL
		  )
		  AND NOT EXISTS (
		    SELECT 1 FROM posts rp
		    WHERE rp.repost_of_id = p.id AND rp.author_agent_id IS NOT NULL
		  )
		  AND NOT EXISTS (
		    SELECT 1 FROM reactions rx
		    WHERE rx.post_id = p.id AND rx.reactor_agent_id IS NOT NULL
		  )
		ORDER BY p.created_at ASC
		LIMIT 50`

	GetRandomRecentPostByOtherAgent = `
		SELECT p.id, p.content, p.author_agent_id, a.handle
		FROM posts p
		JOIN agents a ON p.author_agent_id = a.id
		WHERE p.author_agent_id IS NOT NULL
		  AND p.author_agent_id != $1
		  AND p.created_at > now() - interval '24 hours'
		ORDER BY random()
		LIMIT 1`
)

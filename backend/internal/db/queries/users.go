package queries

const (
	userSelectFull = `
		u.id, u.email, u.handle, u.display_name, u.avatar_url, u.bio, u.website_url, u.is_verified,
		(SELECT COUNT(*) FROM follows  WHERE followee_user_id = u.id)::int AS follower_count,
		(SELECT COUNT(*) FROM follows  WHERE follower_user_id = u.id)::int AS following_count,
		(SELECT COUNT(*) FROM posts    WHERE author_user_id   = u.id)::int AS post_count,
		u.created_at`

	GetUserByID = `SELECT ` + userSelectFull + ` FROM users u WHERE u.id = $1`

	GetUserByHandle = `SELECT ` + userSelectFull + ` FROM users u WHERE u.handle = $1`

	UpsertUser = `
		INSERT INTO users (id, email, handle, display_name, avatar_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email
		RETURNING id, email, handle, display_name, avatar_url, bio, website_url, is_verified,
		          0::int, 0::int, 0::int, created_at`

	UpdateUserProfile = `
		UPDATE users AS u SET
			display_name = COALESCE(NULLIF($2, ''), display_name),
			bio          = CASE WHEN $3::text IS NOT NULL THEN $3 ELSE bio END,
			website_url  = CASE WHEN $4::text IS NOT NULL THEN $4 ELSE website_url END
		WHERE u.id = $1
		RETURNING ` + userSelectFull
)

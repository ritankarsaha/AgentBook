package queries

const (
	GetUserByID = `
		SELECT id, email, handle, display_name, avatar_url, bio, is_verified, created_at
		FROM users WHERE id = $1`


	UpsertUser = `
		INSERT INTO users (id, email, handle, display_name, avatar_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email
		RETURNING id, email, handle, display_name, avatar_url, bio, is_verified, created_at`
)

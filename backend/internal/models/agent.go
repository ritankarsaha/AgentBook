package models

import "time"

type Agent struct {
	ID                string     `json:"id" db:"id"`
	OwnerUserID       string     `json:"owner_user_id" db:"owner_user_id"`
	Handle            string     `json:"handle" db:"handle"`
	DisplayName       string     `json:"display_name" db:"display_name"`
	Description       *string    `json:"description,omitempty" db:"description"`
	Model             string     `json:"model" db:"model"`
	Framework         *string    `json:"framework,omitempty" db:"framework"`
	APIKeyHash        string     `json:"-" db:"api_key_hash"`
	IsVerified        bool       `json:"is_verified" db:"is_verified"`
	VerificationBadge *string    `json:"verification_badge,omitempty" db:"verification_badge"`
	AvatarURL         *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	WebsiteURL        *string    `json:"website_url,omitempty" db:"website_url"`
	AgentReplayID     *string    `json:"agentreplay_id,omitempty" db:"agentreplay_id"`
	LastActiveAt      *time.Time `json:"last_active_at,omitempty" db:"last_active_at"`
	PostCount         int        `json:"post_count" db:"post_count"`
	FollowerCount     int        `json:"follower_count" db:"follower_count"`
	FollowingCount    int        `json:"following_count" db:"following_count"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	Capabilities      []string   `json:"capabilities,omitempty" db:"-"`
}

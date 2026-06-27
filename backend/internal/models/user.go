package models

import "time"

type User struct {
	ID             string    `json:"id" db:"id"`
	Email          string    `json:"email" db:"email"`
	Handle         string    `json:"handle" db:"handle"`
	DisplayName    string    `json:"display_name" db:"display_name"`
	AvatarURL      *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	Bio            *string   `json:"bio,omitempty" db:"bio"`
	WebsiteURL     *string   `json:"website_url,omitempty" db:"website_url"`
	IsVerified     bool      `json:"is_verified" db:"is_verified"`
	FollowerCount  int       `json:"follower_count" db:"follower_count"`
	FollowingCount int       `json:"following_count" db:"following_count"`
	PostCount      int       `json:"post_count" db:"post_count"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

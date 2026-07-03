package models

import "time"

type Post struct {
	ID              string    `json:"id" db:"id"`
	PosterType      string    `json:"poster_type" db:"poster_type"`
	Content         string    `json:"content" db:"content"`
	ReplyToID       *string   `json:"reply_to_id,omitempty" db:"reply_to_id"`
	RepostOfID      *string   `json:"repost_of_id,omitempty" db:"repost_of_id"`
	QuoteContent    *string   `json:"quote_content,omitempty" db:"quote_content"`
	MediaURLs       []string  `json:"media_urls,omitempty" db:"media_urls"`
	PostSubtype     string    `json:"post_subtype" db:"post_subtype"`
	TraceURL        *string   `json:"trace_url,omitempty" db:"trace_url"`
	LikeCount       int       `json:"like_count" db:"like_count"`
	ReplyCount      int       `json:"reply_count" db:"reply_count"`
	RepostCount     int       `json:"repost_count" db:"repost_count"`
	EngagementScore float64   `json:"engagement_score" db:"engagement_score"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`

	AuthorHandle      string  `json:"author_handle" db:"author_handle"`
	AuthorDisplayName string  `json:"author_display_name" db:"author_display_name"`
	AuthorAvatarURL   *string `json:"author_avatar_url,omitempty" db:"author_avatar_url"`
	AuthorIsVerified  bool    `json:"author_is_verified" db:"author_is_verified"`

	// Populated when repost_of_id is set (the original post) or reply_to_id is
	// set (the parent post being replied to) — never both.
	ReferencedPost *Post `json:"referenced_post,omitempty" db:"-"`
}

package handlers

import (
	"time"

	"github.com/ritankar/agentthreads/internal/models"
)

// refPostFields holds nullable columns produced by the LEFT JOIN on posts —
// the original post when repost_of_id is set, or the parent post when
// reply_to_id is set (never both, enforced at creation time). All fields are
// pointers so they scan as nil when the JOIN has no match.
type refPostFields struct {
	ID                *string
	PosterType        *string
	Content           *string
	ReplyToID         *string
	RepostOfID        *string
	QuoteContent      *string
	MediaURLs         []string // nil when no match
	PostSubtype       *string
	TraceURL          *string
	LikeCount         *int
	ReplyCount        *int
	RepostCount       *int
	EngagementScore   *float64
	CreatedAt         *time.Time
	AuthorHandle      *string
	AuthorDisplayName *string
	AuthorAvatarURL   *string
	AuthorIsVerified  *bool
}

// dest returns scan destinations in the same order as the ref-post SELECT columns.
func (r *refPostFields) dest() []any {
	return []any{
		&r.ID, &r.PosterType, &r.Content, &r.ReplyToID, &r.RepostOfID, &r.QuoteContent,
		&r.MediaURLs, &r.PostSubtype, &r.TraceURL, &r.LikeCount, &r.ReplyCount, &r.RepostCount,
		&r.EngagementScore, &r.CreatedAt,
		&r.AuthorHandle, &r.AuthorDisplayName, &r.AuthorAvatarURL, &r.AuthorIsVerified,
	}
}

// toPost converts nullable fields into a *models.Post, or nil if no ref post existed.
func (r *refPostFields) toPost() *models.Post {
	if r.ID == nil {
		return nil
	}
	return &models.Post{
		ID:                *r.ID,
		PosterType:        rpS(r.PosterType),
		Content:           rpS(r.Content),
		ReplyToID:         r.ReplyToID,
		RepostOfID:        r.RepostOfID,
		QuoteContent:      r.QuoteContent,
		MediaURLs:         r.MediaURLs,
		PostSubtype:       rpS(r.PostSubtype),
		TraceURL:          r.TraceURL,
		LikeCount:         rpI(r.LikeCount),
		ReplyCount:        rpI(r.ReplyCount),
		RepostCount:       rpI(r.RepostCount),
		EngagementScore:   rpF(r.EngagementScore),
		CreatedAt:         rpT(r.CreatedAt),
		AuthorHandle:      rpS(r.AuthorHandle),
		AuthorDisplayName: rpS(r.AuthorDisplayName),
		AuthorAvatarURL:   r.AuthorAvatarURL,
		AuthorIsVerified:  rpB(r.AuthorIsVerified),
	}
}

func rpS(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
func rpI(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}
func rpF(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}
func rpB(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
func rpT(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

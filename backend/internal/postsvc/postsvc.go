
package postsvc

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/models"
)

var ErrInvalidParent = errors.New("reply_to_id or repost_of_id does not reference an existing post")

type CreateParams struct {
	AuthorUserID  *string
	AuthorAgentID *string
	PosterType    string
	Content       string
	ReplyToID     *string
	RepostOfID    *string
	QuoteContent  *string
	MediaURLs     []string
	PostSubtype   string
	TraceURL      *string

	AuthorHandle      string
	AuthorDisplayName string
	AuthorAvatarURL   *string
	AuthorIsVerified  bool
}

func Create(ctx context.Context, pool *pgxpool.Pool, params CreateParams) (models.Post, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return models.Post{}, err
	}
	defer tx.Rollback(ctx)

	var post models.Post
	err = tx.QueryRow(ctx, queries.InsertPost,
		params.AuthorUserID, params.AuthorAgentID, params.PosterType, params.Content,
		params.ReplyToID, params.RepostOfID, params.QuoteContent,
		params.MediaURLs, params.PostSubtype, params.TraceURL,
	).Scan(
		&post.ID, &post.PosterType, &post.Content, &post.ReplyToID, &post.RepostOfID, &post.QuoteContent,
		&post.MediaURLs, &post.PostSubtype, &post.TraceURL, &post.LikeCount, &post.ReplyCount, &post.RepostCount,
		&post.EngagementScore, &post.CreatedAt,
	)
	if err != nil {
		if isForeignKeyViolation(err) {
			return models.Post{}, ErrInvalidParent
		}
		return models.Post{}, err
	}

	if params.ReplyToID != nil {
		if _, err := tx.Exec(ctx, queries.IncrementParentReplyCount, *params.ReplyToID); err != nil {
			return models.Post{}, err
		}
	}

	if params.RepostOfID != nil {
		if _, err := tx.Exec(ctx, queries.IncrementParentRepostCount, *params.RepostOfID); err != nil {
			return models.Post{}, err
		}
	}

	if params.AuthorAgentID != nil {
		if _, err := tx.Exec(ctx, queries.IncrementAgentPostCount, *params.AuthorAgentID); err != nil {
			return models.Post{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Post{}, err
	}

	post.AuthorHandle = params.AuthorHandle
	post.AuthorDisplayName = params.AuthorDisplayName
	post.AuthorAvatarURL = params.AuthorAvatarURL
	post.AuthorIsVerified = params.AuthorIsVerified

	return post, nil
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

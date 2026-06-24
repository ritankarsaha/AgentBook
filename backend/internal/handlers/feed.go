package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/models"
)

type FeedHandlers struct {
	Pool *pgxpool.Pool
}

var validPosterTypes = map[string]bool{"all": true, "agent": true, "human": true}
var validFeedSorts = map[string]bool{"new": true, "trending": true}


func (h *FeedHandlers) GetFeed(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	posterType := q.Get("poster_type")
	if posterType == "" {
		posterType = "all"
	}
	if !validPosterTypes[posterType] {
		WriteError(w, http.StatusBadRequest, "poster_type must be one of: all, agent, human")
		return
	}

	sort := q.Get("sort")
	if sort == "" {
		sort = "new"
	}
	if !validFeedSorts[sort] {
		WriteError(w, http.StatusBadRequest, "sort must be one of: new, trending")
		return
	}

	var postSubtype *string
	if s := q.Get("post_subtype"); s != "" {
		postSubtype = &s
	}

	limit := 20
	if l := q.Get("limit"); l != "" {
		n, err := strconv.Atoi(l)
		if err != nil || n < 1 || n > 50 {
			WriteError(w, http.StatusBadRequest, "limit must be an integer between 1 and 50")
			return
		}
		limit = n
	}

	ctx := r.Context()
	cursor := q.Get("cursor")

	if sort == "trending" {
		var scoreCursor *float64
		var idCursor *string
		if cursor != "" {
			s, id, ok := decodeTrendingCursor(cursor)
			if !ok {
				WriteError(w, http.StatusBadRequest, "invalid cursor for sort=trending")
				return
			}
			scoreCursor, idCursor = &s, &id
		}
		rows, err := h.Pool.Query(ctx, queries.FeedTrending, posterType, postSubtype, scoreCursor, idCursor, limit)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "query failed")
			return
		}
		defer rows.Close()
		posts, nextCursor, err := scanTrendingRows(rows)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "scan failed")
			return
		}
		WriteOKWithCursor(w, posts, nextCursor)
		return
	}

	var createdAtCursor *time.Time
	var idCursor *string
	if cursor != "" {
		t, id, ok := decodeNewCursor(cursor)
		if !ok {
			WriteError(w, http.StatusBadRequest, "invalid cursor for sort=new")
			return
		}
		createdAtCursor, idCursor = &t, &id
	}
	rows, err := h.Pool.Query(ctx, queries.FeedPublic, posterType, postSubtype, createdAtCursor, idCursor, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()
	posts, nextCursor, err := scanFeedRows(rows)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "scan failed")
		return
	}
	WriteOKWithCursor(w, posts, nextCursor)
}

func scanFeedRows(rows pgx.Rows) ([]models.Post, string, error) {
	var posts []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(
			&p.ID, &p.PosterType, &p.Content, &p.ReplyToID, &p.RepostOfID, &p.QuoteContent,
			&p.MediaURLs, &p.PostSubtype, &p.TraceURL, &p.LikeCount, &p.ReplyCount, &p.RepostCount,
			&p.EngagementScore, &p.CreatedAt,
			&p.AuthorHandle, &p.AuthorDisplayName, &p.AuthorAvatarURL, &p.AuthorIsVerified,
		); err != nil {
			return nil, "", err
		}
		posts = append(posts, p)
	}
	var next string
	if len(posts) > 0 {
		last := posts[len(posts)-1]
		next = encodeNewCursor(last.CreatedAt, last.ID)
	}
	return posts, next, rows.Err()
}

func scanTrendingRows(rows pgx.Rows) ([]models.Post, string, error) {
	var posts []models.Post
	var lastScore float64
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(
			&p.ID, &p.PosterType, &p.Content, &p.ReplyToID, &p.RepostOfID, &p.QuoteContent,
			&p.MediaURLs, &p.PostSubtype, &p.TraceURL, &p.LikeCount, &p.ReplyCount, &p.RepostCount,
			&p.EngagementScore, &p.CreatedAt,
			&p.AuthorHandle, &p.AuthorDisplayName, &p.AuthorAvatarURL, &p.AuthorIsVerified,
			&lastScore,
		); err != nil {
			return nil, "", err
		}
		posts = append(posts, p)
	}
	var next string
	if len(posts) > 0 {
		next = encodeTrendingCursor(lastScore, posts[len(posts)-1].ID)
	}
	return posts, next, rows.Err()
}

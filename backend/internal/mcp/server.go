// Package mcp implements the MCP Streamable HTTP transport for AgentThreads.
// It exposes six tools for AI agents: post_content, read_feed, search, follow,
// get_agent_profile, and get_notifications.
//
// Auth: write tools require an agent API key via Authorization: Bearer header.
// Read tools (read_feed, search, get_agent_profile) are public.
// The server runs on MCP_PORT (default 8081), separate from the REST API.
package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/middleware"
	"github.com/ritankar/agentthreads/internal/models"
	"github.com/ritankar/agentthreads/internal/postsvc"
)

// ─── JSON-RPC 2.0 wire types ────────────────────────────────────────────────

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"` // string | float64 | nil (MCP allows all three)
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type initializeResult struct {
	ProtocolVersion string `json:"protocolVersion"`
	Capabilities    struct {
		Tools *struct{} `json:"tools"`
	} `json:"capabilities"`
	ServerInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

type toolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema toolInputSchema `json:"inputSchema"`
}

type toolInputSchema struct {
	Type       string                `json:"type"`
	Properties map[string]schemaProp `json:"properties,omitempty"`
	Required   []string              `json:"required,omitempty"`
}

type schemaProp struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type textContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// toolCallResult is the MCP-spec CallToolResult.
type toolCallResult struct {
	Content []textContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ─── Server ──────────────────────────────────────────────────────────────────

// Server is the MCP HTTP server. It runs on a dedicated port alongside the
// REST API server (both in the same process).
type Server struct {
	pool *pgxpool.Pool
	addr string
}

// New constructs a Server that will listen on :<port>.
func New(pool *pgxpool.Pool, port string) *Server {
	return &Server{pool: pool, addr: ":" + port}
}

// Start begins serving MCP requests and blocks until ctx is cancelled.
// Call as a goroutine: go server.Start(ctx).
func (s *Server) Start(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.dispatch)

	srv := &http.Server{Addr: s.addr, Handler: mux}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	log.Printf("mcp: listening on %s", s.addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("mcp: server stopped: %v", err)
	}
}

// dispatch is the single HTTP handler for /mcp.
func (s *Server) dispatch(w http.ResponseWriter, r *http.Request) {
	// CORS — MCP clients (Claude Desktop, Claude Code) may run in browsers
	// or from localhost; allow all origins on the dedicated MCP port.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req rpcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRPCError(w, nil, -32700, "parse error")
		return
	}

	switch req.Method {
	case "initialize":
		s.handleInitialize(w, req)
	case "initialized":
		// Client notification — no response body, just 202.
		w.WriteHeader(http.StatusAccepted)
	case "ping":
		writeRPCResult(w, req.ID, map[string]any{})
	case "tools/list":
		s.handleToolsList(w, req)
	case "tools/call":
		s.handleToolsCall(w, r, req)
	default:
		writeRPCError(w, req.ID, -32601, "method not found: "+req.Method)
	}
}

func (s *Server) handleInitialize(w http.ResponseWriter, req rpcRequest) {
	var r initializeResult
	r.ProtocolVersion = "2024-11-05"
	r.Capabilities.Tools = &struct{}{}
	r.ServerInfo.Name = "AgentThreads"
	r.ServerInfo.Version = "1.0.0"
	writeRPCResult(w, req.ID, r)
}

func (s *Server) handleToolsList(w http.ResponseWriter, req rpcRequest) {
	writeRPCResult(w, req.ID, map[string]any{"tools": toolDefinitions()})
}

func (s *Server) handleToolsCall(w http.ResponseWriter, r *http.Request, req rpcRequest) {
	var params toolCallParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			writeRPCError(w, req.ID, -32602, "invalid params")
			return
		}
	}
	if params.Arguments == nil {
		params.Arguments = map[string]any{}
	}

	ctx := r.Context()

	// Resolve agent from Bearer token when present.
	var agent *models.Agent
	if tok := bearerToken(r); tok != "" {
		if a, err := middleware.ResolveAgent(ctx, s.pool, tok); err == nil {
			agent = a
		}
	}

	var (
		result toolCallResult
		err    error
	)

	switch params.Name {
	// ── Write tools — agent auth required ──────────────────────────
	case "post_content":
		if agent == nil {
			writeRPCResult(w, req.ID, errResult("authentication required: include your agent API key as Authorization: Bearer <key>"))
			return
		}
		result, err = s.toolPostContent(ctx, agent, params.Arguments)

	case "follow":
		if agent == nil {
			writeRPCResult(w, req.ID, errResult("authentication required: include your agent API key as Authorization: Bearer <key>"))
			return
		}
		result, err = s.toolFollow(ctx, agent, params.Arguments)

	case "get_notifications":
		if agent == nil {
			writeRPCResult(w, req.ID, errResult("authentication required: include your agent API key as Authorization: Bearer <key>"))
			return
		}
		result, err = s.toolGetNotifications(ctx, agent, params.Arguments)

	// ── Read tools — public (no auth required) ──────────────────────
	case "read_feed":
		result, err = s.toolReadFeed(ctx, params.Arguments)
	case "search":
		result, err = s.toolSearch(ctx, params.Arguments)
	case "get_agent_profile":
		result, err = s.toolGetAgentProfile(ctx, params.Arguments)

	default:
		writeRPCError(w, req.ID, -32602, "unknown tool: "+params.Name)
		return
	}

	if err != nil {
		writeRPCResult(w, req.ID, errResult("internal error: "+err.Error()))
		return
	}
	writeRPCResult(w, req.ID, result)
}

// ─── Tool implementations ─────────────────────────────────────────────────────

func (s *Server) toolPostContent(ctx context.Context, agent *models.Agent, args map[string]any) (toolCallResult, error) {
	content, _ := args["content"].(string)
	if content == "" {
		return errResult("content is required"), nil
	}
	if len([]rune(content)) > 500 {
		return errResult("content exceeds 500 characters"), nil
	}

	params := postsvc.CreateParams{
		AuthorAgentID:     &agent.ID,
		PosterType:        "agent",
		Content:           content,
		PostSubtype:       "standard",
		AuthorHandle:      agent.Handle,
		AuthorDisplayName: agent.DisplayName,
		AuthorAvatarURL:   agent.AvatarURL,
		AuthorIsVerified:  agent.IsVerified,
	}
	if v, ok := args["reply_to_id"].(string); ok && v != "" {
		params.ReplyToID = &v
	}
	if v, ok := args["post_subtype"].(string); ok && v != "" {
		valid := map[string]bool{"standard": true, "trace": true, "task-log": true, "finding": true}
		if !valid[v] {
			return errResult("post_subtype must be one of: standard, trace, task-log, finding"), nil
		}
		params.PostSubtype = v
	}
	if v, ok := args["trace_url"].(string); ok && v != "" {
		params.TraceURL = &v
	}

	post, err := postsvc.Create(ctx, s.pool, params)
	if err != nil {
		if err == postsvc.ErrInvalidParent {
			return errResult("reply_to_id references a post that does not exist"), nil
		}
		return toolCallResult{}, fmt.Errorf("postsvc.Create: %w", err)
	}

	out, _ := json.MarshalIndent(post, "", "  ")
	return okResult(fmt.Sprintf("Post created successfully (id: %s)\n\n%s", post.ID, string(out))), nil
}

func (s *Server) toolReadFeed(ctx context.Context, args map[string]any) (toolCallResult, error) {
	posterType, _ := args["poster_type"].(string)
	if posterType == "" {
		posterType = "all"
	}
	switch posterType {
	case "all", "agent", "human":
	default:
		return errResult("poster_type must be one of: all, agent, human"), nil
	}

	sort, _ := args["sort"].(string)
	if sort == "" {
		sort = "new"
	}
	switch sort {
	case "new", "trending":
	default:
		return errResult("sort must be one of: new, trending"), nil
	}

	limit := 20
	if v, ok := args["limit"].(float64); ok && v >= 1 && v <= 50 {
		limit = int(v)
	}

	cursor, _ := args["cursor"].(string)

	var posts []models.Post
	var nextCursor string

	if sort == "trending" {
		var scoreCursor *float64
		var idCursor *string
		if cursor != "" {
			sc, id, ok := decodeTrendingCursor(cursor)
			if !ok {
				return errResult("invalid cursor for sort=trending"), nil
			}
			scoreCursor, idCursor = &sc, &id
		}
		rows, err := s.pool.Query(ctx, queries.FeedTrending, posterType, nil, scoreCursor, idCursor, limit)
		if err != nil {
			return toolCallResult{}, fmt.Errorf("feed query: %w", err)
		}
		defer rows.Close()
		posts, nextCursor, err = scanTrendingFeedRows(rows)
		if err != nil {
			return toolCallResult{}, fmt.Errorf("scan: %w", err)
		}
	} else {
		var createdAtCursor *time.Time
		var idCursor *string
		if cursor != "" {
			t, id, ok := decodeNewCursor(cursor)
			if !ok {
				return errResult("invalid cursor for sort=new"), nil
			}
			createdAtCursor, idCursor = &t, &id
		}
		rows, err := s.pool.Query(ctx, queries.FeedPublic, posterType, nil, createdAtCursor, idCursor, limit)
		if err != nil {
			return toolCallResult{}, fmt.Errorf("feed query: %w", err)
		}
		defer rows.Close()
		posts, nextCursor, err = scanNewFeedRows(rows)
		if err != nil {
			return toolCallResult{}, fmt.Errorf("scan: %w", err)
		}
	}

	if posts == nil {
		posts = []models.Post{}
	}
	out, _ := json.MarshalIndent(map[string]any{
		"posts":       posts,
		"count":       len(posts),
		"next_cursor": nextCursor,
	}, "", "  ")
	return okResult(string(out)), nil
}

func (s *Server) toolSearch(ctx context.Context, args map[string]any) (toolCallResult, error) {
	q, _ := args["query"].(string)
	if q == "" {
		return errResult("query is required"), nil
	}
	if len(q) < 2 {
		return errResult("query must be at least 2 characters"), nil
	}

	searchType, _ := args["type"].(string)
	if searchType == "" {
		searchType = "all"
	}
	switch searchType {
	case "posts", "agents", "all":
	default:
		return errResult("type must be one of: posts, agents, all"), nil
	}

	var capability *string
	if v, ok := args["capability"].(string); ok && v != "" {
		capability = &v
	}

	limit := 20
	if v, ok := args["limit"].(float64); ok && v >= 1 && v <= 50 {
		limit = int(v)
	}

	type result struct {
		Posts  []models.Post  `json:"posts,omitempty"`
		Agents []models.Agent `json:"agents,omitempty"`
	}
	res := result{}

	if searchType == "posts" || searchType == "all" {
		rows, err := s.pool.Query(ctx, queries.SearchPosts, q, limit)
		if err != nil {
			return toolCallResult{}, fmt.Errorf("search posts: %w", err)
		}
		for rows.Next() {
			var p models.Post
			if err := rows.Scan(
				&p.ID, &p.PosterType, &p.Content, &p.ReplyToID, &p.RepostOfID, &p.QuoteContent,
				&p.MediaURLs, &p.PostSubtype, &p.TraceURL, &p.LikeCount, &p.ReplyCount,
				&p.RepostCount, &p.EngagementScore, &p.CreatedAt,
				&p.AuthorHandle, &p.AuthorDisplayName, &p.AuthorAvatarURL, &p.AuthorIsVerified,
			); err != nil {
				continue
			}
			if p.MediaURLs == nil {
				p.MediaURLs = []string{}
			}
			res.Posts = append(res.Posts, p)
		}
		rows.Close()
	}

	if searchType == "agents" || searchType == "all" {
		rows, err := s.pool.Query(ctx, queries.SearchAgents, q, capability, limit)
		if err != nil {
			return toolCallResult{}, fmt.Errorf("search agents: %w", err)
		}
		for rows.Next() {
			var a models.Agent
			if err := rows.Scan(
				&a.ID, &a.OwnerUserID, &a.Handle, &a.DisplayName, &a.Description,
				&a.Model, &a.Framework, &a.APIKeyHash,
				&a.IsVerified, &a.VerificationBadge, &a.AvatarURL, &a.WebsiteURL,
				&a.AgentReplayID, &a.LastActiveAt, &a.PostCount, &a.FollowerCount,
				&a.FollowingCount, &a.CreatedAt,
			); err != nil {
				continue
			}
			a.APIKeyHash = ""
			res.Agents = append(res.Agents, a)
		}
		rows.Close()
	}

	if res.Posts == nil {
		res.Posts = []models.Post{}
	}
	if res.Agents == nil {
		res.Agents = []models.Agent{}
	}
	out, _ := json.MarshalIndent(res, "", "  ")
	return okResult(string(out)), nil
}

func (s *Server) toolFollow(ctx context.Context, agent *models.Agent, args map[string]any) (toolCallResult, error) {
	handle, _ := args["handle"].(string)
	if handle == "" {
		return errResult("handle is required"), nil
	}
	handle = strings.TrimPrefix(handle, "@")

	followeeID, followeeType, err := resolveHandle(ctx, s.pool, handle)
	if err != nil {
		return errResult("no user or agent found with handle: @" + handle), nil
	}
	if followeeType == "agent" && followeeID == agent.ID {
		return errResult("you cannot follow yourself"), nil
	}

	agentID := agent.ID
	var followeeUserID, followeeAgentID *string
	if followeeType == "agent" {
		followeeAgentID = &followeeID
	} else {
		followeeUserID = &followeeID
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return toolCallResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, queries.InsertFollow,
		nil, &agentID, "agent",
		followeeUserID, followeeAgentID, followeeType,
	)
	if err != nil {
		return toolCallResult{}, fmt.Errorf("insert follow: %w", err)
	}

	if tag.RowsAffected() > 0 {
		if followeeType == "agent" {
			_, _ = tx.Exec(ctx, queries.IncrementAgentFollowerCount, followeeID)
		}
		_, _ = tx.Exec(ctx, queries.IncrementAgentFollowingCount, agentID)
	}

	if err := tx.Commit(ctx); err != nil {
		return toolCallResult{}, fmt.Errorf("commit: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return okResult("Already following @" + handle + "."), nil
	}
	return okResult("Now following @" + handle + "."), nil
}

func (s *Server) toolGetAgentProfile(ctx context.Context, args map[string]any) (toolCallResult, error) {
	handle, _ := args["handle"].(string)
	if handle == "" {
		return errResult("handle is required"), nil
	}
	handle = strings.TrimPrefix(handle, "@")

	var a models.Agent
	err := s.pool.QueryRow(ctx, queries.GetAgentByHandle, handle).Scan(
		&a.ID, &a.OwnerUserID, &a.Handle, &a.DisplayName, &a.Description,
		&a.Model, &a.Framework, &a.APIKeyHash, &a.IsVerified, &a.VerificationBadge,
		&a.AvatarURL, &a.WebsiteURL, &a.AgentReplayID, &a.LastActiveAt,
		&a.PostCount, &a.FollowerCount, &a.FollowingCount, &a.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errResult("no agent found with handle: @" + handle), nil
		}
		return toolCallResult{}, fmt.Errorf("get agent: %w", err)
	}
	a.APIKeyHash = ""

	capRows, err := s.pool.Query(ctx, queries.GetAgentCapabilities, a.ID)
	if err == nil {
		for capRows.Next() {
			var cap string
			if capRows.Scan(&cap) == nil {
				a.Capabilities = append(a.Capabilities, cap)
			}
		}
		capRows.Close()
	}

	out, _ := json.MarshalIndent(a, "", "  ")
	return okResult(string(out)), nil
}

func (s *Server) toolGetNotifications(ctx context.Context, agent *models.Agent, args map[string]any) (toolCallResult, error) {
	limit := 30
	if v, ok := args["limit"].(float64); ok && v >= 1 && v <= 100 {
		limit = int(v)
	}

	var cursorTime *time.Time
	if v, ok := args["cursor"].(string); ok && v != "" {
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			cursorTime = &t
		}
	}

	agentID := agent.ID
	rows, err := s.pool.Query(ctx, queries.GetNotifications, nil, &agentID, cursorTime, limit)
	if err != nil {
		return toolCallResult{}, fmt.Errorf("get notifications: %w", err)
	}

	type notif struct {
		ID               string    `json:"id"`
		Type             string    `json:"type"`
		ActorHandle      *string   `json:"actor_handle,omitempty"`
		ActorDisplayName *string   `json:"actor_display_name,omitempty"`
		PostID           *string   `json:"post_id,omitempty"`
		Read             bool      `json:"read"`
		CreatedAt        time.Time `json:"created_at"`
	}

	notifs := []notif{}
	var lastCreatedAt *time.Time
	for rows.Next() {
		var n notif
		var actorUserID, actorAgentID, actorAvatarURL *string
		if err := rows.Scan(
			&n.ID, &n.Type, &actorUserID, &actorAgentID, &n.PostID, &n.Read, &n.CreatedAt,
			&n.ActorHandle, &n.ActorDisplayName, &actorAvatarURL,
		); err != nil {
			continue
		}
		notifs = append(notifs, n)
		t := n.CreatedAt
		lastCreatedAt = &t
	}
	rows.Close()

	markRead := true
	if v, ok := args["mark_read"].(bool); ok {
		markRead = v
	}
	if markRead && len(notifs) > 0 {
		go func() {
			_, _ = s.pool.Exec(context.Background(), queries.MarkNotificationsRead, nil, &agentID)
		}()
	}

	var nextCursor string
	if len(notifs) == limit && lastCreatedAt != nil {
		nextCursor = lastCreatedAt.Format(time.RFC3339Nano)
	}

	out, _ := json.MarshalIndent(map[string]any{
		"notifications": notifs,
		"count":         len(notifs),
		"next_cursor":   nextCursor,
	}, "", "  ")
	return okResult(string(out)), nil
}

// ─── Tool definitions (for tools/list) ───────────────────────────────────────

func toolDefinitions() []toolDef {
	return []toolDef{
		{
			Name:        "post_content",
			Description: "Create a post on AgentThreads. The post appears immediately in the public feed. Requires agent authentication.",
			InputSchema: toolInputSchema{
				Type: "object",
				Properties: map[string]schemaProp{
					"content": {
						Type:        "string",
						Description: "Post text (max 500 characters). Emoji and personality welcome.",
					},
					"reply_to_id": {
						Type:        "string",
						Description: "UUID of the post this is a reply to (optional).",
					},
					"post_subtype": {
						Type:        "string",
						Description: "Post type: standard (default), trace (links to an AgentReplay trace), task-log, or finding.",
						Enum:        []string{"standard", "trace", "task-log", "finding"},
					},
					"trace_url": {
						Type:        "string",
						Description: "URL to an AgentReplay trace (required when post_subtype=trace).",
					},
				},
				Required: []string{"content"},
			},
		},
		{
			Name:        "read_feed",
			Description: "Read the public AgentThreads feed. Returns a paginated list of posts from agents and humans. No authentication required.",
			InputSchema: toolInputSchema{
				Type: "object",
				Properties: map[string]schemaProp{
					"poster_type": {
						Type:        "string",
						Description: "Filter by poster type. Default: all.",
						Enum:        []string{"all", "agent", "human"},
					},
					"sort": {
						Type:        "string",
						Description: "Feed sort order. 'new' (default) returns newest first; 'trending' ranks by engagement.",
						Enum:        []string{"new", "trending"},
					},
					"limit": {
						Type:        "number",
						Description: "Number of posts to return (1–50). Default: 20.",
					},
					"cursor": {
						Type:        "string",
						Description: "Pagination cursor from a previous read_feed response.",
					},
				},
			},
		},
		{
			Name:        "search",
			Description: "Full-text search over posts and agent profiles using Postgres FTS. No authentication required.",
			InputSchema: toolInputSchema{
				Type: "object",
				Properties: map[string]schemaProp{
					"query": {
						Type:        "string",
						Description: "Search query (minimum 2 characters).",
					},
					"type": {
						Type:        "string",
						Description: "What to search: posts, agents, or all (default).",
						Enum:        []string{"posts", "agents", "all"},
					},
					"capability": {
						Type:        "string",
						Description: "Filter agent results by capability tag (e.g. 'rust', 'trading').",
					},
					"limit": {
						Type:        "number",
						Description: "Max results per type (1–50). Default: 20.",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "follow",
			Description: "Follow an agent or human by handle. Idempotent (safe to call if already following). Requires agent authentication.",
			InputSchema: toolInputSchema{
				Type: "object",
				Properties: map[string]schemaProp{
					"handle": {
						Type:        "string",
						Description: "Handle to follow (with or without leading @).",
					},
				},
				Required: []string{"handle"},
			},
		},
		{
			Name:        "get_agent_profile",
			Description: "Fetch a full agent profile including capabilities, post count, and follower/following counts. No authentication required.",
			InputSchema: toolInputSchema{
				Type: "object",
				Properties: map[string]schemaProp{
					"handle": {
						Type:        "string",
						Description: "Agent handle to look up (with or without leading @).",
					},
				},
				Required: []string{"handle"},
			},
		},
		{
			Name:        "get_notifications",
			Description: "Fetch notifications for the authenticated agent (follows, likes, replies, reposts). Requires agent authentication.",
			InputSchema: toolInputSchema{
				Type: "object",
				Properties: map[string]schemaProp{
					"limit": {
						Type:        "number",
						Description: "Number of notifications to return (1–100). Default: 30.",
					},
					"cursor": {
						Type:        "string",
						Description: "Pagination cursor (RFC3339Nano timestamp) from a previous call.",
					},
					"mark_read": {
						Type:        "string",
						Description: "Set to 'false' to skip marking returned notifications as read. Default: marks read.",
					},
				},
			},
		},
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func okResult(text string) toolCallResult {
	return toolCallResult{Content: []textContent{{Type: "text", Text: text}}}
}

func errResult(text string) toolCallResult {
	return toolCallResult{Content: []textContent{{Type: "text", Text: text}}, IsError: true}
}

func writeRPCResult(w http.ResponseWriter, id any, result any) {
	json.NewEncoder(w).Encode(rpcResponse{JSONRPC: "2.0", ID: id, Result: result})
}

func writeRPCError(w http.ResponseWriter, id any, code int, msg string) {
	json.NewEncoder(w).Encode(rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}})
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(h, "Bearer ")
}

// resolveHandle looks up a handle in the agents+users shared namespace and
// returns the entity's ID and type ("agent" or "human").
func resolveHandle(ctx context.Context, pool *pgxpool.Pool, handle string) (id, entityType string, err error) {
	err = pool.QueryRow(ctx, queries.GetFolloweeByHandle, handle).Scan(&id, &entityType)
	if err == pgx.ErrNoRows {
		return "", "", pgx.ErrNoRows
	}
	return id, entityType, err
}

// ─── Cursor codec (mirrors handlers/cursor.go — kept local to avoid coupling) ─

func encodeNewCursor(t time.Time, id string) string {
	raw := t.UTC().Format(time.RFC3339Nano) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeNewCursor(cursor string) (time.Time, string, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", false
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", false
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", false
	}
	return t, parts[1], true
}

func encodeTrendingCursor(score float64, id string) string {
	raw := strconv.FormatFloat(score, 'f', -1, 64) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeTrendingCursor(cursor string) (float64, string, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, "", false
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 {
		return 0, "", false
	}
	score, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, "", false
	}
	return score, parts[1], true
}

// ─── Feed row scanners (mirrors handlers/feed.go — local to avoid coupling) ──

func scanNewFeedRows(rows pgx.Rows) ([]models.Post, string, error) {
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
		if p.MediaURLs == nil {
			p.MediaURLs = []string{}
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

func scanTrendingFeedRows(rows pgx.Rows) ([]models.Post, string, error) {
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
		if p.MediaURLs == nil {
			p.MediaURLs = []string{}
		}
		posts = append(posts, p)
	}
	var next string
	if len(posts) > 0 {
		next = encodeTrendingCursor(lastScore, posts[len(posts)-1].ID)
	}
	return posts, next, rows.Err()
}

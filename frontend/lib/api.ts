const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function getServerToken(): Promise<string | null> {
  try {
    const res = await fetch("/api/auth/token", { credentials: "same-origin" });
    if (!res.ok) return null;
    const json = await res.json();
    return json.access_token ?? null;
  } catch {
    return null;
  }
}

export type PosterType = "human" | "agent";
export type PostSubtype = "standard" | "trace" | "task-log" | "finding";

export type User = {
  id: string;
  email: string;
  handle: string;
  display_name: string;
  avatar_url?: string;
  bio?: string;
  website_url?: string;
  is_verified: boolean;
  follower_count: number;
  following_count: number;
  post_count: number;
  created_at: string;
};

export type Agent = {
  id: string;
  handle: string;
  display_name: string;
  description?: string;
  model: string;
  framework?: string;
  is_verified: boolean;
  verification_badge?: string;
  avatar_url?: string;
  website_url?: string;
  agentreplay_id?: string;
  follower_count: number;
  following_count: number;
  post_count: number;
  capabilities?: string[];
};

export type Post = {
  id: string;
  poster_type: PosterType;
  content: string;
  reply_to_id?: string;
  repost_of_id?: string;
  quote_content?: string;
  media_urls?: string[];
  post_subtype: PostSubtype;
  trace_url?: string;
  like_count: number;
  reply_count: number;
  repost_count: number;
  engagement_score: number;
  created_at: string;
  author_handle: string;
  author_display_name: string;
  author_avatar_url?: string;
  author_is_verified: boolean;
};

type ApiEnvelope<T> = {
  ok: boolean;
  data: T | null;
  cursor?: string;
  error: string | null;
};

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<ApiEnvelope<T>> {
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    cache: "no-store",
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
  });

  const body = (await res.json()) as ApiEnvelope<T>;

  if (!res.ok || !body.ok) {
    throw new ApiError(body.error ?? `request failed with status ${res.status}`, res.status);
  }

  return body;
}

function buildQuery(params?: Record<string, string | number | undefined>): string {
  if (!params) return "";
  const entries = Object.entries(params).filter(
    ([, v]) => v !== undefined && v !== ""
  ) as [string, string | number][];
  if (entries.length === 0) return "";
  return "?" + new URLSearchParams(entries.map(([k, v]) => [k, String(v)])).toString();
}

export function apiGet<T>(
  path: string,
  params?: Record<string, string | number | undefined>,
  token?: string
): Promise<ApiEnvelope<T>> {
  return request<T>(`${path}${buildQuery(params)}`, {
    method: "GET",
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  });
}

export function apiPost<T>(path: string, body?: unknown, token?: string): Promise<ApiEnvelope<T>> {
  return request<T>(path, {
    method: "POST",
    body: body !== undefined ? JSON.stringify(body) : undefined,
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  });
}

export function apiPut<T>(path: string, body?: unknown, token?: string): Promise<ApiEnvelope<T>> {
  return request<T>(path, {
    method: "PUT",
    body: body !== undefined ? JSON.stringify(body) : undefined,
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  });
}

export function apiDelete<T>(path: string, token?: string): Promise<ApiEnvelope<T>> {
  return request<T>(path, {
    method: "DELETE",
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  });
}

export function likePost(postId: string, token: string): Promise<ApiEnvelope<{ liked: boolean }>> {
  return apiPost<{ liked: boolean }>(`/api/v1/reactions/${postId}`, undefined, token);
}

export function unlikePost(postId: string, token: string): Promise<ApiEnvelope<{ liked: boolean }>> {
  return apiDelete<{ liked: boolean }>(`/api/v1/reactions/${postId}`, token);
}

export function repostPost(postId: string, token: string): Promise<ApiEnvelope<Post>> {
  return apiPost<Post>("/api/v1/posts", { repost_of_id: postId }, token);
}

export function quoteRepostPost(
  postId: string,
  quoteContent: string,
  token: string
): Promise<ApiEnvelope<Post>> {
  return apiPost<Post>("/api/v1/posts", { repost_of_id: postId, quote_content: quoteContent }, token);
}

export type RegisterAgentRequest = {
  handle: string;
  display_name: string;
  description?: string;
  model: string;
  framework?: string;
  website_url?: string;
  agentreplay_id?: string;
  capabilities?: string[];
};

export type RegisterAgentResponse = {
  agent: Agent;
  api_key: string;
};

export function registerAgent(
  req: RegisterAgentRequest,
  token: string
): Promise<ApiEnvelope<RegisterAgentResponse>> {
  return apiPost<RegisterAgentResponse>("/api/v1/agents/register", req, token);
}

export function listMyAgents(token: string): Promise<ApiEnvelope<(Agent & { capabilities: string[] })[]>> {
  return apiGet<(Agent & { capabilities: string[] })[]>("/api/v1/users/me/agents", undefined, token);
}

export function checkHandleAvailable(handle: string): Promise<ApiEnvelope<Agent>> {
  return apiGet<Agent>(`/api/v1/agents/${encodeURIComponent(handle)}`);
}

export type CaptchaPuzzle = {
  token: string;
  type: "base64_compute" | "json_path" | "bitwise";
  question: string;
  expires_at: string;
};

export type VerifyAgentResponse = {
  verified: boolean;
  badge: string;
};

/** Request a reverse-CAPTCHA puzzle for the calling agent. Requires the agent's API key. */
export function getVerifyPuzzle(
  agentApiKey: string
): Promise<ApiEnvelope<CaptchaPuzzle>> {
  return apiGet<CaptchaPuzzle>("/api/v1/agents/me/verify", undefined, agentApiKey);
}

/** Submit a puzzle answer. Requires the agent's API key. */
export function verifyAgent(
  token: string,
  answer: string,
  agentApiKey: string
): Promise<ApiEnvelope<VerifyAgentResponse>> {
  return apiPost<VerifyAgentResponse>(
    "/api/v1/agents/me/verify",
    { token, answer },
    agentApiKey
  );
}

// ─── Social graph ─────────────────────────────────────────────────────────────

export function followHandle(
  handle: string,
  token: string
): Promise<ApiEnvelope<{ following: boolean }>> {
  return apiPost<{ following: boolean }>(
    `/api/v1/follows/${encodeURIComponent(handle)}`,
    undefined,
    token
  );
}

export function unfollowHandle(
  handle: string,
  token: string
): Promise<ApiEnvelope<{ following: boolean }>> {
  return apiDelete<{ following: boolean }>(
    `/api/v1/follows/${encodeURIComponent(handle)}`,
    token
  );
}

// ─── Search ───────────────────────────────────────────────────────────────────

export type SearchResult = {
  posts: Post[];
  agents: Agent[];
};

export function search(
  q: string,
  type: "posts" | "agents" | "all" = "all",
  capability?: string,
  limit?: number
): Promise<ApiEnvelope<SearchResult>> {
  return apiGet<SearchResult>("/api/v1/search", { q, type, capability, limit });
}

// ─── Notifications ────────────────────────────────────────────────────────────

export type Notification = {
  id: string;
  type: string;
  actor_handle?: string;
  actor_display_name?: string;
  actor_avatar_url?: string;
  post_id?: string;
  read: boolean;
  created_at: string;
};

export function getNotifications(
  token: string,
  cursor?: string,
  limit?: number
): Promise<ApiEnvelope<Notification[]>> {
  return apiGet<Notification[]>("/api/v1/notifications", { cursor, limit }, token);
}

// ─── Stats ────────────────────────────────────────────────────────────────────

export type PlatformStats = {
  registered_agents: number;
  active_agents_7d: number;
  human_users: number;
  posts_total: number;
};

export function getPlatformStats(): Promise<ApiEnvelope<PlatformStats>> {
  return apiGet<PlatformStats>("/api/v1/stats");
}

// ─── Media ────────────────────────────────────────────────────────────────────

export type UploadURLResponse = {
  signed_url: string;
  path: string;
  expires_at: string;
};

export function getMediaUploadURL(
  contentType: string,
  filename: string,
  token: string
): Promise<ApiEnvelope<UploadURLResponse>> {
  return apiPost<UploadURLResponse>(
    "/api/v1/media/upload-url",
    { content_type: contentType, filename },
    token
  );
}

export function getFollowingFeed(
  token: string,
  cursor?: string,
  limit = 20
): Promise<ApiEnvelope<Post[]>> {
  return apiGet<Post[]>("/api/v1/feed/following", { cursor, limit }, token);
}

// ─── Profiles ─────────────────────────────────────────────────────────────────

export type AgentProfile = Agent & {
  capabilities: string[];
  owner_user_id: string;
  description?: string;
  framework?: string;
  website_url?: string;
  agentreplay_id?: string;
  last_active_at?: string;
};

export function getAgentProfile(handle: string): Promise<ApiEnvelope<AgentProfile>> {
  return apiGet<AgentProfile>(`/api/v1/agents/${encodeURIComponent(handle)}`);
}

export function getUserProfile(handle: string): Promise<ApiEnvelope<User>> {
  return apiGet<User>(`/api/v1/users/${encodeURIComponent(handle)}`);
}

export type UpdateUserRequest = {
  display_name?: string;
  bio?: string;
  website_url?: string;
};

export function updateUserProfile(
  req: UpdateUserRequest,
  token: string,
): Promise<ApiEnvelope<User>> {
  return apiPut<User>("/api/v1/users/me", req, token);
}

export type ProfileTab = "posts" | "replies" | "traces";

export function getAgentPosts(
  handle: string,
  tab: ProfileTab = "posts",
  cursor?: string,
  limit = 20
): Promise<ApiEnvelope<Post[]>> {
  return apiGet<Post[]>(`/api/v1/agents/${encodeURIComponent(handle)}/posts`, {
    tab,
    cursor,
    limit,
  });
}

export function getUserPosts(
  handle: string,
  tab: ProfileTab = "posts",
  cursor?: string,
  limit = 20
): Promise<ApiEnvelope<Post[]>> {
  return apiGet<Post[]>(`/api/v1/users/${encodeURIComponent(handle)}/posts`, {
    tab,
    cursor,
    limit,
  });
}

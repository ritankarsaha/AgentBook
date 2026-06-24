const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export type PosterType = "human" | "agent";
export type PostSubtype = "standard" | "trace" | "task-log" | "finding";

export type User = {
  id: string;
  email: string;
  handle: string;
  display_name: string;
  avatar_url?: string;
  bio?: string;
  is_verified: boolean;
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

export function apiDelete<T>(path: string, token?: string): Promise<ApiEnvelope<T>> {
  return request<T>(path, {
    method: "DELETE",
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  });
}

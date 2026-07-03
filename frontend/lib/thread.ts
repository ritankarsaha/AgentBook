import { cache } from "react";
import { notFound } from "next/navigation";
import { apiGet, ApiError, type Post } from "@/lib/api";

export type ThreadResponse = { post: Post; replies: Post[] };

// Same request-scoped dedup rationale as lib/profile.ts's resolveProfile —
// apiGet() opts out of Next's fetch memoization via `cache: "no-store"`, so
// this is what keeps generateMetadata() and the page component from each
// hitting the Go backend separately for the same thread.
export const resolveThread = cache(async (id: string): Promise<ThreadResponse> => {
  try {
    const res = await apiGet<ThreadResponse>(`/api/v1/posts/${id}`);
    if (!res.data) notFound();
    return res.data;
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound();
    throw err;
  }
});

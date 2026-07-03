import { cache } from "react";
import { notFound } from "next/navigation";
import { getAgentProfile, getUserProfile, type AgentProfile, type User } from "@/lib/api";

export type ResolvedProfile =
  | { type: "agent"; agent: AgentProfile }
  | { type: "user"; user: User };

// Wrapped in React's cache() so the page component and generateMetadata()
// share one resolution per request instead of each hitting the Go backend
// separately — apiGet() sets `cache: "no-store"` on its fetch, which opts
// out of Next's own request memoization, so this is the only dedup we get.
export const resolveProfile = cache(async (handle: string): Promise<ResolvedProfile> => {
  try {
    const res = await getAgentProfile(handle);
    if (res.data) return { type: "agent", agent: res.data };
  } catch {
    // Any agent-lookup failure (404 or otherwise) falls through to try the
    // handle as a human user next — matches the platform's original behavior.
  }

  try {
    const res = await getUserProfile(handle);
    if (res.data) return { type: "user", user: res.data };
  } catch {
    // Falls through to notFound() below.
  }

  notFound();
});

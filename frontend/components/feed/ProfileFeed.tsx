"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { PostCard } from "./PostCard";
import { AgentReplayCard } from "../agent/AgentReplayCard";
import { getAgentPosts, getUserPosts } from "@/lib/api";
import type { Post, ProfileTab } from "@/lib/api";

const ALL_TABS: { label: string; value: ProfileTab }[] = [
  { label: "Posts", value: "posts" },
  { label: "Replies", value: "replies" },
  { label: "Traces", value: "traces" },
];

interface ProfileFeedProps {
  handle: string;
  profileType: "agent" | "user";
  initialPosts: Post[];
  initialCursor: string;
  tab: ProfileTab;
  availableTabs?: ProfileTab[];
}

export function ProfileFeed({
  handle,
  profileType,
  initialPosts,
  initialCursor,
  tab,
  availableTabs,
}: ProfileFeedProps) {
  const tabs = availableTabs
    ? ALL_TABS.filter((t) => availableTabs.includes(t.value))
    : ALL_TABS;
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const [posts, setPosts] = useState<Post[]>(initialPosts);
  const [cursor, setCursor] = useState(initialCursor);
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(!initialCursor);
  const sentinel = useRef<HTMLDivElement>(null);

  const loadMore = useCallback(async () => {
    if (loading || done || !cursor) return;
    setLoading(true);
    try {
      const fn = profileType === "agent" ? getAgentPosts : getUserPosts;
      const res = await fn(handle, tab, cursor);
      const next = res.data ?? [];
      setPosts((prev) => [...prev, ...next]);
      setCursor(res.cursor ?? "");
      if (!res.cursor) setDone(true);
    } catch {
      setDone(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, done, handle, loading, profileType, tab]);

  useEffect(() => {
    const el = sentinel.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) loadMore();
      },
      { rootMargin: "200px" }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [loadMore]);

  function switchTab(t: ProfileTab) {
    const params = new URLSearchParams(searchParams.toString());
    params.set("tab", t);
    router.push(`${pathname}?${params.toString()}`);
  }

  return (
    <div>
      {/* Tab bar */}
      <div className="flex border-b border-border">
        {tabs.map(({ label, value }) => (
          <button
            key={value}
            type="button"
            onClick={() => switchTab(value)}
            className={`flex-1 py-3 text-sm font-medium transition-colors ${
              tab === value
                ? "border-b-2 border-accent text-accent"
                : "text-text-muted hover:text-text-secondary"
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Posts */}
      <div className="divide-y divide-border">
        {posts.length === 0 && !loading && (
          <p className="py-10 text-center text-sm text-text-muted">No posts here yet.</p>
        )}
        {posts.map((post) =>
          post.post_subtype === "trace" ? (
            <div key={post.id} className="p-4">
              <AgentReplayCard post={post} />
            </div>
          ) : (
            <PostCard key={post.id} post={post} />
          )
        )}
      </div>

      <div ref={sentinel} className="h-4" />
      {loading && <p className="py-6 text-center text-sm text-text-muted">Loading…</p>}
      {done && posts.length > 0 && (
        <p className="py-6 text-center text-xs text-text-muted">All posts loaded.</p>
      )}
    </div>
  );
}

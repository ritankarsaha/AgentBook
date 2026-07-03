"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { apiGet, type Post, type PosterType, type PostSubtype, type User } from "@/lib/api";
import { createClient } from "@/lib/supabase/client";
import { PostCard } from "./PostCard";
import { PostComposer } from "./PostComposer";
import { EmptyState } from "@/components/ui/EmptyState";

type FeedParams = {
  posterType?: PosterType | "all";
  postSubtype?: PostSubtype;
  sort?: "new" | "trending";
};

type RawPostRow = {
  id: string;
  poster_type: PosterType;
  post_subtype: PostSubtype;
};

export function InfiniteFeed({
  initialPosts,
  initialCursor,
  params = {},
  composerUser,
}: {
  initialPosts: Post[];
  initialCursor: string;
  params?: FeedParams;
  composerUser?: User | null;
}) {
  const [posts, setPosts] = useState(initialPosts);
  const [cursor, setCursor] = useState(initialCursor);
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(!initialCursor);
  const [pending, setPending] = useState<Post[]>([]);
  const sentinelRef = useRef<HTMLDivElement>(null);
  const knownIds = useRef(new Set(initialPosts.map((p) => p.id)));

  const loadMore = useCallback(async () => {
    if (loading || done || !cursor) return;
    setLoading(true);
    try {
      const res = await apiGet<Post[]>("/api/v1/feed", {
        poster_type: params.posterType,
        post_subtype: params.postSubtype,
        sort: params.sort,
        cursor,
        limit: 20,
      });
      const next = res.data ?? [];
      next.forEach((p) => knownIds.current.add(p.id));
      setPosts((prev) => [...prev, ...next]);
      if (!res.cursor || next.length === 0) {
        setDone(true);
      } else {
        setCursor(res.cursor);
      }
    } catch {
      setDone(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, done, loading, params.posterType, params.postSubtype, params.sort]);

  useEffect(() => {
    const el = sentinelRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) {
          loadMore();
        }
      },
      { rootMargin: "400px" }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [loadMore]);

  useEffect(() => {
    if (params.sort === "trending") return;

    const supabase = createClient();
    const channel = supabase
      .channel("posts-feed")
      .on(
        "postgres_changes",
        { event: "INSERT", schema: "public", table: "posts" },
        async ({ new: row }: { new: RawPostRow }) => {
          if (knownIds.current.has(row.id)) return;
          if (params.posterType && params.posterType !== "all" && row.poster_type !== params.posterType) return;
          if (params.postSubtype && row.post_subtype !== params.postSubtype) return;

          knownIds.current.add(row.id);
          try {
            const res = await apiGet<{ post: Post }>(`/api/v1/posts/${row.id}`);
            if (res.data?.post) {
              setPending((prev) => [...prev, res.data!.post]);
            }
          } catch {

          }
        }
      )
      .subscribe();

    return () => {
      supabase.removeChannel(channel);
    };
  }, [params.sort, params.posterType, params.postSubtype]);

  function showPending() {
    setPosts((prev) => [...pending].reverse().concat(prev));
    setPending([]);
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function handlePosted(post: Post) {
    knownIds.current.add(post.id);
    setPosts((prev) => [post, ...prev]);
  }

  return (
    <div>
      {composerUser && <PostComposer user={composerUser} onPosted={handlePosted} />}
      {pending.length > 0 && (
        <button
          onClick={showPending}
          className="block w-full border-b border-border bg-accent/10 py-2.5 text-center text-sm font-medium text-accent hover:bg-accent/20"
        >
          Show {pending.length} new post{pending.length > 1 ? "s" : ""}
        </button>
      )}
      {posts.map((post) => (
        <PostCard key={post.id} post={post} user={composerUser} />
      ))}
      {!done && <div ref={sentinelRef} className="h-10" />}
      {loading && <p className="py-4 text-center text-sm text-text-muted">Loading…</p>}
      {done && posts.length === 0 && (
        <EmptyState
          title="No posts yet"
          subtitle="Follow some agents to fill this feed with their posts."
          ctaHref="/explore"
          ctaLabel="Explore Agents →"
        />
      )}
    </div>
  );
}

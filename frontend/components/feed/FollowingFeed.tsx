"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import Link from "next/link";
import { getFollowingFeed, getServerToken } from "@/lib/api";
import { PostCard } from "./PostCard";
import type { Post } from "@/lib/api";

export function FollowingFeed() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [cursor, setCursor] = useState("");
  const [loading, setLoading] = useState(true);
  const [done, setDone] = useState(false);
  const [authed, setAuthed] = useState<boolean | null>(null); // null = unknown
  const tokenRef = useRef<string | null>(null);
  const sentinel = useRef<HTMLDivElement>(null);

  // Resolve token once on mount via server-side route (avoids stale browser cookie issues).
  useEffect(() => {
    getServerToken().then((t) => {
      tokenRef.current = t;
      setAuthed(t !== null);
      if (t === null) setLoading(false);
    });
  }, []);

  // Fetch first page once we know we're authed.
  useEffect(() => {
    if (!authed || !tokenRef.current) return;
    setLoading(true);
    getFollowingFeed(tokenRef.current, undefined, 20)
      .then((res) => {
        setPosts(res.data ?? []);
        setCursor(res.cursor ?? "");
        if (!res.cursor) setDone(true);
      })
      .catch(() => setDone(true))
      .finally(() => setLoading(false));
  }, [authed]);

  const loadMore = useCallback(async () => {
    if (loading || done || !cursor || !tokenRef.current) return;
    setLoading(true);
    try {
      const res = await getFollowingFeed(tokenRef.current, cursor, 20);
      const next = res.data ?? [];
      setPosts((prev) => [...prev, ...next]);
      setCursor(res.cursor ?? "");
      if (!res.cursor) setDone(true);
    } catch {
      setDone(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, done, loading]);

  useEffect(() => {
    const el = sentinel.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      (entries) => { if (entries[0].isIntersecting) loadMore(); },
      { rootMargin: "400px" }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [loadMore]);

  if (authed === null || (authed && loading && posts.length === 0)) {
    return <p className="py-10 text-center text-sm text-text-muted">Loading…</p>;
  }

  if (!authed) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-center">
        <p className="text-text-secondary">Sign in to see posts from accounts you follow.</p>
        <Link
          href="/login"
          className="rounded-full bg-accent px-5 py-2 text-sm font-medium text-white hover:opacity-90"
        >
          Sign in
        </Link>
      </div>
    );
  }

  if (!loading && posts.length === 0) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-center">
        <p className="text-text-secondary">Nothing here yet.</p>
        <p className="text-sm text-text-muted">Follow some agents or humans to see their posts.</p>
        <Link
          href="/explore"
          className="rounded-full border border-border px-4 py-1.5 text-sm text-text-secondary hover:border-accent hover:text-accent"
        >
          Explore agents →
        </Link>
      </div>
    );
  }

  return (
    <div className="divide-y divide-border">
      {posts.map((post) => <PostCard key={post.id} post={post} />)}
      <div ref={sentinel} className="h-4" />
      {loading && <p className="py-4 text-center text-sm text-text-muted">Loading…</p>}
      {done && posts.length > 0 && (
        <p className="py-6 text-center text-xs text-text-muted">All caught up.</p>
      )}
    </div>
  );
}

"use client";

import Image from "next/image";
import Link from "next/link";
import { useState } from "react";
import { likePost, quoteRepostPost, repostPost, unlikePost, type Post } from "@/lib/api";
import { createClient } from "@/lib/supabase/client";

const QUOTE_MAX_LEN = 500;

function timeAgo(iso: string): string {
  const diffSec = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (diffSec < 60) return `${Math.max(diffSec, 0)}s`;
  const min = Math.floor(diffSec / 60);
  if (min < 60) return `${min}m`;
  const hr = Math.floor(min / 60);
  if (hr < 24) return `${hr}h`;
  return `${Math.floor(hr / 24)}d`;
}

async function getToken(): Promise<string | null> {
  const supabase = createClient();
  const { data } = await supabase.auth.getSession();
  return data.session?.access_token ?? null;
}

export function PostCard({ post }: { post: Post }) {
  const isAgent = post.poster_type === "agent";
  const handleColorClass = isAgent ? "text-accent-agent" : "text-accent-human";
  const borderColorClass = isAgent ? "border-l-accent-agent" : "border-l-accent-human";

  const [liked, setLiked] = useState(false);
  const [likeCount, setLikeCount] = useState(post.like_count);
  const [reposted, setReposted] = useState(false);
  const [repostCount, setRepostCount] = useState(post.repost_count);
  const [busy, setBusy] = useState(false);
  const [quoteOpen, setQuoteOpen] = useState(false);
  const [quoteText, setQuoteText] = useState("");
  const [quoteSubmitting, setQuoteSubmitting] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  async function toggleLike() {
    if (busy) return;
    const token = await getToken();
    if (!token) {
      setActionError("Sign in to like posts.");
      return;
    }
    setBusy(true);
    setActionError(null);
    try {
      if (liked) {
        await unlikePost(post.id, token);
        setLiked(false);
        setLikeCount((c) => Math.max(c - 1, 0));
      } else {
        await likePost(post.id, token);
        setLiked(true);
        setLikeCount((c) => c + 1);
      }
    } catch {
      setActionError("Couldn't update like — try again.");
    } finally {
      setBusy(false);
    }
  }

  async function doRepost() {
    if (busy || reposted) return;
    const token = await getToken();
    if (!token) {
      setActionError("Sign in to repost.");
      return;
    }
    setBusy(true);
    setActionError(null);
    try {
      await repostPost(post.id, token);
      setReposted(true);
      setRepostCount((c) => c + 1);
    } catch {
      setActionError("Couldn't repost — try again.");
    } finally {
      setBusy(false);
    }
  }

  async function submitQuote() {
    const trimmed = quoteText.trim();
    if (!trimmed || trimmed.length > QUOTE_MAX_LEN || quoteSubmitting) return;
    const token = await getToken();
    if (!token) {
      setActionError("Sign in to quote-repost.");
      return;
    }
    setQuoteSubmitting(true);
    setActionError(null);
    try {
      await quoteRepostPost(post.id, trimmed, token);
      setReposted(true);
      setRepostCount((c) => c + 1);
      setQuoteText("");
      setQuoteOpen(false);
    } catch {
      setActionError("Couldn't quote-repost — try again.");
    } finally {
      setQuoteSubmitting(false);
    }
  }

  return (
    <article className={`flex gap-3 border-b border-border bg-surface px-4 py-3 border-l-2 ${borderColorClass}`}>
      <div className="h-10 w-10 shrink-0 overflow-hidden rounded-full bg-border">
        {post.author_avatar_url ? (
          <Image
            src={post.author_avatar_url}
            alt={post.author_handle}
            width={40}
            height={40}
            className="h-10 w-10 object-cover"
          />
        ) : (
          <div className="flex h-10 w-10 items-center justify-center font-mono text-sm text-text-secondary">
            {post.author_display_name.charAt(0).toUpperCase()}
          </div>
        )}
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-baseline gap-x-2">
          <Link href={`/${post.author_handle}`} className={`font-mono font-medium ${handleColorClass}`}>
            @{post.author_handle}
          </Link>
          {isAgent && (
            <span className="text-xs text-text-muted">
              🤖 agent{post.author_is_verified ? " · verified ✓" : ""}
            </span>
          )}
          <span className="text-sm text-text-secondary">{post.author_display_name}</span>
          {/* timeAgo() is intentionally a few seconds off between SSR and hydration */}
          <Link
            href={`/posts/${post.id}`}
            className="text-sm text-text-muted hover:underline"
            suppressHydrationWarning
          >
            · {timeAgo(post.created_at)}
          </Link>
        </div>

        {post.quote_content && (
          <p className="mt-1 whitespace-pre-wrap break-words text-[15px] leading-relaxed text-text-primary">
            {post.quote_content}
          </p>
        )}

        {post.content && (
          <p className="mt-1 whitespace-pre-wrap break-words text-[15px] leading-relaxed text-text-primary">
            {post.content}
          </p>
        )}

        {post.repost_of_id && (
          <Link
            href={`/posts/${post.repost_of_id}`}
            className="mt-1 inline-block text-sm text-text-muted hover:underline"
          >
            🔁 {post.quote_content ? "Quoted a post" : "Reposted a post"} — view original
          </Link>
        )}

        {post.post_subtype === "trace" && post.trace_url && (
          <a
            href={post.trace_url}
            target="_blank"
            rel="noreferrer"
            className="mt-2 inline-block text-sm text-accent hover:underline"
          >
            → View AgentReplay Trace
          </a>
        )}

        <div className="mt-2 flex items-center gap-6 text-sm text-text-muted">
          <button
            type="button"
            onClick={toggleLike}
            disabled={busy}
            aria-pressed={liked}
            className={`transition-colors hover:text-danger disabled:cursor-not-allowed ${liked ? "text-danger" : ""}`}
          >
            {liked ? "❤️" : "🤍"} {likeCount}
          </button>
          <Link href={`/posts/${post.id}`} className="hover:text-text-secondary">
            💬 {post.reply_count}
          </Link>
          <button
            type="button"
            onClick={doRepost}
            disabled={busy || reposted}
            aria-pressed={reposted}
            className={`transition-colors hover:text-accent disabled:cursor-not-allowed ${reposted ? "text-accent" : ""}`}
          >
            🔁 {repostCount}
          </button>
          <button
            type="button"
            onClick={() => setQuoteOpen((open) => !open)}
            className="hover:text-text-secondary"
          >
            Quote
          </button>
        </div>

        {actionError && (
          <p className="mt-1 text-sm text-danger" role="alert">
            {actionError}
          </p>
        )}

        {quoteOpen && (
          <div className="mt-2 rounded-xl border border-border bg-bg p-2">
            <textarea
              value={quoteText}
              onChange={(e) => setQuoteText(e.target.value)}
              placeholder="Add a comment…"
              rows={2}
              maxLength={QUOTE_MAX_LEN}
              className="w-full resize-none bg-transparent text-[15px] text-text-primary placeholder:text-text-muted focus:outline-none"
            />
            <div className="mt-1 flex items-center justify-between">
              <span className="font-mono text-xs text-text-muted">
                {QUOTE_MAX_LEN - quoteText.length}
              </span>
              <button
                type="button"
                onClick={submitQuote}
                disabled={!quoteText.trim() || quoteSubmitting}
                className="rounded-full bg-accent px-3 py-1 text-xs font-medium text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
              >
                {quoteSubmitting ? "Posting…" : "Quote-repost"}
              </button>
            </div>
          </div>
        )}
      </div>
    </article>
  );
}

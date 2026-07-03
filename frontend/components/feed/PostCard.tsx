"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { getServerToken, likePost, repostPost, unlikePost, type Post, type User } from "@/lib/api";
import { EmbeddedPostCard } from "./EmbeddedPostCard";
import { ReplyModal } from "./ReplyModal";
import { QuoteModal } from "./QuoteModal";

function timeAgo(iso: string): string {
  const diffSec = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (diffSec < 60) return `${Math.max(diffSec, 0)}s`;
  const min = Math.floor(diffSec / 60);
  if (min < 60) return `${min}m`;
  const hr = Math.floor(min / 60);
  if (hr < 24) return `${hr}h`;
  return `${Math.floor(hr / 24)}d`;
}

function Avatar({ url, name, size = 40 }: { url?: string; name: string; size?: number }) {
  return (
    <div
      className="shrink-0 overflow-hidden rounded-full bg-border"
      style={{ width: size, height: size }}
    >
      {url ? (
        <Image
          src={url}
          alt={name}
          width={size}
          height={size}
          className="object-cover"
          style={{ width: size, height: size }}
        />
      ) : (
        <div
          className="flex items-center justify-center font-mono text-text-secondary"
          style={{ width: size, height: size, fontSize: size * 0.35 }}
        >
          {name.charAt(0).toUpperCase()}
        </div>
      )}
    </div>
  );
}

// ─── Repost dropdown ─────────────────────────────────────────────────────────

function RepostMenu({
  user,
  reposted,
  repostCount,
  busy,
  onRepost,
  onQuoteOpen,
}: {
  user?: User | null;
  reposted: boolean;
  repostCount: number;
  busy: boolean;
  onRepost: () => void;
  onQuoteOpen: () => void;
}) {
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    function onClickOutside(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", onClickOutside);
    return () => document.removeEventListener("mousedown", onClickOutside);
  }, [open]);

  return (
    <div className="relative" ref={menuRef}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        disabled={busy}
        aria-label="Repost options"
        className={`flex items-center gap-1.5 rounded-full px-2 py-1.5 text-sm transition-colors hover:bg-accent/10 hover:text-accent disabled:cursor-not-allowed ${
          reposted ? "text-accent" : "text-text-muted"
        }`}
      >
        {/* Repost icon */}
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="17 1 21 5 17 9" />
          <path d="M3 11V9a4 4 0 0 1 4-4h14" />
          <polyline points="7 23 3 19 7 15" />
          <path d="M21 13v2a4 4 0 0 1-4 4H3" />
        </svg>
        <span>{repostCount}</span>
      </button>

      {open && (
        <div className="absolute bottom-full left-1/2 mb-2 -translate-x-1/2 z-40 min-w-[160px] rounded-xl border border-border bg-surface shadow-xl overflow-hidden">
          <button
            type="button"
            onClick={() => { setOpen(false); onRepost(); }}
            disabled={reposted || busy || !user}
            className="flex w-full items-center gap-3 px-4 py-3 text-sm font-medium text-text-primary transition-colors hover:bg-border disabled:cursor-not-allowed disabled:opacity-50"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="17 1 21 5 17 9" />
              <path d="M3 11V9a4 4 0 0 1 4-4h14" />
              <polyline points="7 23 3 19 7 15" />
              <path d="M21 13v2a4 4 0 0 1-4 4H3" />
            </svg>
            {reposted ? "Reposted" : "Repost"}
          </button>
          <div className="h-px bg-border" />
          <button
            type="button"
            onClick={() => { setOpen(false); onQuoteOpen(); }}
            disabled={!user}
            className="flex w-full items-center gap-3 px-4 py-3 text-sm font-medium text-text-primary transition-colors hover:bg-border disabled:cursor-not-allowed disabled:opacity-50"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
            </svg>
            Quote
          </button>
        </div>
      )}
    </div>
  );
}

// ─── Main PostCard ────────────────────────────────────────────────────────────

type PostCardProps = {
  post: Post;
  user?: User | null;
  // Prevents recursive repost-header rendering inside an already-embedded card.
  isEmbedded?: boolean;
  // Suppresses the "Replying to" preview — used on thread pages, where the
  // parent post this reply targets is already rendered directly above it.
  hideReplyContext?: boolean;
  onReplied?: (reply: Post) => void;
};

export function PostCard({ post, user, isEmbedded = false, hideReplyContext = false, onReplied }: PostCardProps) {
  // Classify post type
  const isPlainRepost = !!post.repost_of_id && !post.quote_content;
  const isQuoteRepost = !!post.repost_of_id && !!post.quote_content;
  const isAgent = post.poster_type === "agent";
  const handleColorClass = isAgent ? "text-accent-agent" : "text-accent-human";
  const borderColorClass = isAgent ? "border-l-accent-agent" : "border-l-accent-human";

  // ── All hooks unconditionally first (Rules of Hooks) ──────────────────────
  const [liked, setLiked] = useState(false);
  const [likeCount, setLikeCount] = useState(post.like_count);
  const [reposted, setReposted] = useState(false);
  const [repostCount, setRepostCount] = useState(post.repost_count);
  const [replyCount, setReplyCount] = useState(post.reply_count);
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);
  const [replyModalOpen, setReplyModalOpen] = useState(false);
  const [quoteModalOpen, setQuoteModalOpen] = useState(false);

  // ── Plain repost: delegate to the referenced post + show reposter header ──
  if (isPlainRepost && !isEmbedded) {
    const original = post.referenced_post;
    return (
      <div className="border-b border-border bg-surface">
        {/* Repost header */}
        <div className="flex items-center gap-2 px-4 pt-4 pb-1.5 text-xs text-text-muted">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="17 1 21 5 17 9" />
            <path d="M3 11V9a4 4 0 0 1 4-4h14" />
            <polyline points="7 23 3 19 7 15" />
            <path d="M21 13v2a4 4 0 0 1-4 4H3" />
          </svg>
          <Link href={`/${post.author_handle}`} className="font-medium hover:underline">
            {post.author_display_name}
          </Link>
          <span>reposted</span>
        </div>
        {/* Render the original post card (isEmbedded prevents another repost header) */}
        {original ? (
          <PostCard post={original} user={user} isEmbedded onReplied={onReplied} />
        ) : (
          <p className="px-4 py-3 text-sm text-text-muted">Original post unavailable.</p>
        )}
      </div>
    );
  }

  // ── Handlers ──────────────────────────────────────────────────────────────
  async function toggleLike() {
    if (busy) return;
    const token = await getServerToken();
    if (!token) { setActionError("Sign in to like posts."); return; }
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
    const token = await getServerToken();
    if (!token) { setActionError("Sign in to repost."); return; }
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

  function handleReplied(reply: Post) {
    setReplyCount((c) => c + 1);
    onReplied?.(reply);
  }

  // What to show as main text content
  const displayContent = isQuoteRepost ? post.quote_content : post.content;

  return (
    <>
      <article
        className={`flex gap-3 border-b border-border bg-surface px-4 py-4 ${
          !isEmbedded ? `border-l-2 ${borderColorClass}` : ""
        }`}
      >
        <Avatar url={post.author_avatar_url} name={post.author_display_name} size={40} />

        <div className="min-w-0 flex-1">
          {/* Reply context: which post this is replying to, with a preview */}
          {post.reply_to_id && !hideReplyContext && (
            <div className="mb-2">
              <p className="text-xs text-text-muted">
                Replying to{" "}
                {post.referenced_post ? (
                  <Link
                    href={`/${post.referenced_post.author_handle}`}
                    className={`font-mono ${
                      post.referenced_post.poster_type === "agent"
                        ? "text-accent-agent"
                        : "text-accent-human"
                    } hover:underline`}
                  >
                    @{post.referenced_post.author_handle}
                  </Link>
                ) : (
                  "a post"
                )}
              </p>
              {post.referenced_post && <EmbeddedPostCard post={post.referenced_post} />}
            </div>
          )}

          {/* Author row */}
          <div className="flex flex-wrap items-baseline gap-x-2">
            <Link
              href={`/${post.author_handle}`}
              className={`font-mono text-sm font-semibold ${handleColorClass} hover:underline`}
            >
              @{post.author_handle}
            </Link>
            {isAgent && (
              <span className="text-xs text-text-muted">
                🤖{post.author_is_verified ? " verified ✓" : " agent"}
              </span>
            )}
            <span className="text-sm text-text-secondary">{post.author_display_name}</span>
            <Link
              href={`/posts/${post.id}`}
              className="text-xs text-text-muted hover:underline"
              suppressHydrationWarning
            >
              · {timeAgo(post.created_at)}
            </Link>
          </div>

          {/* Main content */}
          {displayContent && (
            <p className="mt-1.5 whitespace-pre-wrap break-words text-[15px] leading-relaxed text-text-primary">
              {displayContent}
            </p>
          )}

          {/* Embedded original for quote-reposts */}
          {isQuoteRepost && post.referenced_post && (
            <EmbeddedPostCard post={post.referenced_post} />
          )}
          {isQuoteRepost && !post.referenced_post && post.repost_of_id && (
            <Link
              href={`/posts/${post.repost_of_id}`}
              className="mt-2 inline-block text-xs text-text-muted hover:underline"
            >
              View original post →
            </Link>
          )}

          {/* Trace link */}
          {post.post_subtype === "trace" && post.trace_url && (
            <a
              href={post.trace_url}
              target="_blank"
              rel="noreferrer"
              className="mt-2 inline-flex items-center gap-1 text-sm text-accent hover:underline"
            >
              → View AgentReplay Trace
            </a>
          )}

          {/* Action bar */}
          <div className="mt-3 flex items-center gap-1">
            {/* Reply */}
            <button
              type="button"
              onClick={() => setReplyModalOpen(true)}
              aria-label="Reply"
              className="flex items-center gap-1.5 rounded-full px-2 py-1.5 text-sm text-text-muted transition-colors hover:bg-accent/10 hover:text-accent"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
              </svg>
              <span>{replyCount}</span>
            </button>

            {/* Repost dropdown */}
            <RepostMenu
              user={user}
              reposted={reposted}
              repostCount={repostCount}
              busy={busy}
              onRepost={doRepost}
              onQuoteOpen={() => setQuoteModalOpen(true)}
            />

            {/* Like */}
            <button
              type="button"
              onClick={toggleLike}
              disabled={busy}
              aria-pressed={liked}
              aria-label={liked ? "Unlike" : "Like"}
              className={`flex items-center gap-1.5 rounded-full px-2 py-1.5 text-sm transition-colors hover:bg-danger/10 hover:text-danger disabled:cursor-not-allowed ${
                liked ? "text-danger" : "text-text-muted"
              }`}
            >
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill={liked ? "currentColor" : "none"}
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
              </svg>
              <span>{likeCount}</span>
            </button>

            {/* View thread link */}
            <Link
              href={`/posts/${post.id}`}
              aria-label="View thread"
              className="ml-auto flex items-center rounded-full p-1.5 text-text-muted transition-colors hover:bg-border hover:text-text-secondary"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
                <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
              </svg>
            </Link>
          </div>

          {actionError && (
            <p className="mt-1 text-xs text-danger" role="alert">
              {actionError}
            </p>
          )}
        </div>
      </article>

      {/* Modals (rendered outside the article for correct stacking) */}
      {replyModalOpen && (
        <ReplyModal
          post={post}
          user={user ?? null}
          onClose={() => setReplyModalOpen(false)}
          onReplied={handleReplied}
        />
      )}
      {quoteModalOpen && (
        <QuoteModal
          post={post}
          user={user ?? null}
          onClose={() => setQuoteModalOpen(false)}
          onQuoted={(newPost) => {
            setRepostCount((c) => c + 1);
            setReposted(true);
            onReplied?.(newPost);
          }}
        />
      )}
    </>
  );
}

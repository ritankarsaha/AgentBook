"use client";

import Image from "next/image";
import { useEffect, useRef, useState } from "react";
import { createReply, getServerToken, type Post, type User } from "@/lib/api";

const MAX_LEN = 500;

type Props = {
  post: Post;
  user: User | null;
  onClose: () => void;
  onReplied: (reply: Post) => void;
};

export function ReplyModal({ post, user, onClose, onReplied }: Props) {
  const [content, setContent] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const isAgent = post.poster_type === "agent";
  const handleColor = isAgent ? "text-accent-agent" : "text-accent-human";

  const remaining = MAX_LEN - content.length;
  const counterColor =
    remaining < 0
      ? "text-danger"
      : remaining <= 10
      ? "text-danger"
      : remaining <= 50
      ? "text-amber-400"
      : "text-text-muted";

  const canSubmit = content.trim().length > 0 && content.length <= MAX_LEN && !submitting;

  // Focus textarea on open
  useEffect(() => {
    textareaRef.current?.focus();
  }, []);

  // Esc to close
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  async function submit() {
    if (!canSubmit) return;
    if (!user) {
      setError("Sign in to reply.");
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      const token = await getServerToken();
      if (!token) {
        setError("Session expired — sign in again.");
        return;
      }
      const res = await createReply(post.id, content.trim(), token);
      if (res.data) {
        onReplied(res.data);
        onClose();
      }
    } catch {
      setError("Couldn't post reply — try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    // Overlay
    <div
      className="fixed inset-0 z-50 flex items-start justify-center bg-black/60 backdrop-blur-sm pt-16 px-4"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="w-full max-w-lg rounded-2xl border border-border bg-surface shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-border px-4 py-3">
          <button
            onClick={onClose}
            className="rounded-full p-1.5 text-text-secondary transition-colors hover:bg-border hover:text-text-primary"
            aria-label="Close"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
          <span className="text-sm font-medium text-text-secondary">Reply</span>
          <div className="w-8" />
        </div>

        <div className="px-4 pt-4 pb-2">
          {/* Original post preview */}
          <div className="flex gap-3">
            <div className="flex flex-col items-center">
              <div className="h-9 w-9 shrink-0 overflow-hidden rounded-full bg-border">
                {post.author_avatar_url ? (
                  <Image
                    src={post.author_avatar_url}
                    alt={post.author_handle}
                    width={36}
                    height={36}
                    className="h-9 w-9 object-cover"
                  />
                ) : (
                  <div className="flex h-9 w-9 items-center justify-center font-mono text-sm text-text-secondary">
                    {post.author_display_name.charAt(0).toUpperCase()}
                  </div>
                )}
              </div>
              {/* Vertical connector line */}
              <div className="mt-1 w-0.5 flex-1 bg-border" />
            </div>
            <div className="min-w-0 flex-1 pb-3">
              <div className="flex items-baseline gap-2">
                <span className={`font-mono text-sm font-medium ${handleColor}`}>
                  @{post.author_handle}
                </span>
                <span className="text-sm text-text-secondary">{post.author_display_name}</span>
              </div>
              <p className="mt-1 line-clamp-4 whitespace-pre-wrap text-sm leading-relaxed text-text-primary">
                {post.content || post.quote_content}
              </p>
              <p className="mt-1 text-xs text-text-muted">
                Replying to{" "}
                <span className={`font-mono ${handleColor}`}>@{post.author_handle}</span>
              </p>
            </div>
          </div>

          {/* Reply composer */}
          <div className="flex gap-3">
            <div className="h-9 w-9 shrink-0 overflow-hidden rounded-full bg-border">
              {user?.avatar_url ? (
                <Image
                  src={user.avatar_url}
                  alt={user.handle}
                  width={36}
                  height={36}
                  className="h-9 w-9 object-cover"
                />
              ) : (
                <div className="flex h-9 w-9 items-center justify-center font-mono text-sm text-text-secondary">
                  {user ? user.display_name.charAt(0).toUpperCase() : "?"}
                </div>
              )}
            </div>
            <div className="min-w-0 flex-1">
              <textarea
                ref={textareaRef}
                value={content}
                onChange={(e) => setContent(e.target.value)}
                onKeyDown={(e) => {
                  if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
                    e.preventDefault();
                    submit();
                  }
                }}
                placeholder="Post your reply"
                rows={4}
                className="w-full resize-none bg-transparent text-[15px] text-text-primary placeholder:text-text-muted focus:outline-none"
              />
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between border-t border-border px-4 py-3">
          <span className={`font-mono text-xs ${counterColor}`}>{remaining}</span>
          {error && (
            <p className="text-xs text-danger" role="alert">
              {error}
            </p>
          )}
          <button
            onClick={submit}
            disabled={!canSubmit || !user}
            className="rounded-full bg-accent px-5 py-2 text-sm font-semibold text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
          >
            {submitting ? "Posting…" : "Reply"}
          </button>
        </div>
      </div>
    </div>
  );
}

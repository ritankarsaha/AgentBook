"use client";

import Image from "next/image";
import { useEffect, useRef, useState } from "react";
import { apiPost, getServerToken, type Post, type User } from "@/lib/api";
import { useToast } from "@/components/ui/Toast";

const MAX_LEN = 500;

type Props = {
  user: User;
  onClose: () => void;
  onPosted?: (post: Post) => void;
};

export function ComposeModal({ user, onClose, onPosted }: Props) {
  const [content, setContent] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const toast = useToast();

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

  useEffect(() => {
    textareaRef.current?.focus();
  }, []);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  async function submit() {
    if (!canSubmit) return;
    setSubmitting(true);
    setError(null);
    try {
      const token = await getServerToken();
      if (!token) {
        setError("Session expired — sign in again.");
        toast.error("Session expired — please sign in again.");
        return;
      }
      const res = await apiPost<Post>("/api/v1/posts", { content: content.trim() }, token);
      if (res.data) {
        onPosted?.(res.data);
        toast.success("Post created");
        onClose();
      }
    } catch {
      setError("Couldn't post that — try again.");
      toast.error("Couldn't post that — try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
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
            className="rounded-full p-1.5 text-text-secondary transition-all duration-150 hover:scale-110 hover:bg-border hover:text-text-primary active:scale-90"
            aria-label="Close"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
          <span className="text-sm font-medium text-text-secondary">New post</span>
          <div className="w-8" />
        </div>

        {/* Composer */}
        <div className="px-4 pt-4 pb-2">
          <div className="flex gap-3">
            <div className="h-9 w-9 shrink-0 overflow-hidden rounded-full bg-border">
              {user.avatar_url ? (
                <Image src={user.avatar_url} alt={user.handle} width={36} height={36} className="h-9 w-9 object-cover" />
              ) : (
                <div className="flex h-9 w-9 items-center justify-center font-mono text-sm text-text-secondary">
                  {user.display_name.charAt(0).toUpperCase()}
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
                placeholder="What's happening?"
                rows={5}
                className="w-full resize-none rounded-md bg-transparent text-[15px] text-text-primary placeholder:text-text-muted focus:outline-none focus-visible:ring-2 focus-visible:ring-accent/50"
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
            disabled={!canSubmit}
            className="rounded-full bg-accent px-5 py-2 text-sm font-semibold text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
          >
            {submitting ? "Posting…" : "Post"}
          </button>
        </div>
      </div>
    </div>
  );
}

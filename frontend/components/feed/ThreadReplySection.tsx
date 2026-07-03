"use client";

import Image from "next/image";
import Link from "next/link";
import { useState } from "react";
import { createReply, getServerToken, type Post, type User } from "@/lib/api";
import { PostCard } from "./PostCard";

const MAX_LEN = 500;

type Props = {
  postId: string;
  initialReplies: Post[];
  user: User | null;
  replyTarget: Post;
};

export function ThreadReplySection({ postId, initialReplies, user, replyTarget }: Props) {
  const [replies, setReplies] = useState(initialReplies);
  const [content, setContent] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isAgent = replyTarget.poster_type === "agent";
  const targetHandleColor = isAgent ? "text-accent-agent" : "text-accent-human";

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

  async function submit() {
    if (!canSubmit || !user) return;
    setSubmitting(true);
    setError(null);
    try {
      const token = await getServerToken();
      if (!token) { setError("Session expired."); return; }
      const res = await createReply(postId, content.trim(), token);
      if (res.data) {
        setReplies((prev) => [...prev, res.data!]);
        setContent("");
      }
    } catch {
      setError("Couldn't post reply — try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <>
      {/* Reply composer (logged-in users only) */}
      {user ? (
        <div className="flex gap-3 border-b border-border px-4 py-3">
          <div className="h-9 w-9 shrink-0 overflow-hidden rounded-full bg-border">
            {user.avatar_url ? (
              <Image
                src={user.avatar_url}
                alt={user.handle}
                width={36}
                height={36}
                className="h-9 w-9 object-cover"
              />
            ) : (
              <div className="flex h-9 w-9 items-center justify-center font-mono text-sm text-text-secondary">
                {user.display_name.charAt(0).toUpperCase()}
              </div>
            )}
          </div>
          <div className="min-w-0 flex-1">
            <p className="mb-1 text-xs text-text-muted">
              Replying to{" "}
              <span className={`font-mono ${targetHandleColor}`}>
                @{replyTarget.author_handle}
              </span>
            </p>
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              onKeyDown={(e) => {
                if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
                  e.preventDefault();
                  submit();
                }
              }}
              placeholder="Post your reply"
              rows={3}
              className="w-full resize-none bg-transparent text-[15px] text-text-primary placeholder:text-text-muted focus:outline-none"
            />
            {error && (
              <p className="mt-0.5 text-xs text-danger" role="alert">
                {error}
              </p>
            )}
            <div className="mt-1.5 flex items-center justify-between">
              <span className={`font-mono text-xs ${counterColor}`}>{remaining}</span>
              <button
                onClick={submit}
                disabled={!canSubmit}
                className="rounded-full bg-accent px-4 py-1.5 text-sm font-semibold text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
              >
                {submitting ? "Posting…" : "Reply"}
              </button>
            </div>
          </div>
        </div>
      ) : (
        <div className="border-b border-border px-4 py-4 text-center">
          <p className="text-sm text-text-muted">
            <Link href="/login" className="text-accent hover:underline">
              Sign in
            </Link>{" "}
            to reply
          </p>
        </div>
      )}

      {/* Replies list */}
      {replies.length === 0 ? (
        <p className="py-10 text-center text-text-muted">No replies yet.</p>
      ) : (
        <div>
          {replies.map((reply) => (
            <PostCard key={reply.id} post={reply} user={user} hideReplyContext />
          ))}
        </div>
      )}
    </>
  );
}

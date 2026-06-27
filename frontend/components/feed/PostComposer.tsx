"use client";

import Image from "next/image";
import { useState } from "react";
import { apiPost, getServerToken, type Post, type User } from "@/lib/api";

const MAX_LEN = 500;

// "Post as" only ever shows the signed-in human today — posting as an owned
// agent from the web UI would need (a) a "list my agents" endpoint, which
// doesn't exist (agents/me is agent-API-key auth, not human-JWT auth), and
// (b) a new authorization path letting a human post on behalf of an agent
// they own, which isn't in the API design.
export function PostComposer({
  user,
  onPosted,
}: {
  user: User;
  onPosted: (post: Post) => void;
}) {
  const [content, setContent] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const length = content.length;
  const remaining = MAX_LEN - length;
  const counterColor =
    remaining < 0 ? "text-danger" : remaining <= 10 ? "text-danger" : remaining <= 50 ? "text-amber-400" : "text-text-muted";

  const trimmed = content.trim();
  const canSubmit = trimmed.length > 0 && length <= MAX_LEN && !submitting;

  async function submit() {
    if (!canSubmit) return;
    setSubmitting(true);
    setError(null);
    try {
      const token = await getServerToken();
      if (!token) {
        setError("Session expired — please sign in again.");
        return;
      }
      const res = await apiPost<Post>("/api/v1/posts", { content: trimmed }, token);
      if (res.data) {
        onPosted(res.data);
        setContent("");
      }
    } catch {
      setError("Couldn't post that — try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="flex gap-3 border-b border-border px-4 py-3">
      <div className="h-10 w-10 shrink-0 overflow-hidden rounded-full bg-border">
        {user.avatar_url ? (
          <Image src={user.avatar_url} alt={user.handle} width={40} height={40} className="h-10 w-10 object-cover" />
        ) : (
          <div className="flex h-10 w-10 items-center justify-center font-mono text-sm text-text-secondary">
            {user.display_name.charAt(0).toUpperCase()}
          </div>
        )}
      </div>

      <div className="min-w-0 flex-1">
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          onKeyDown={(e) => {
            if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
              e.preventDefault();
              submit();
            }
          }}
          placeholder="What's happening?"
          rows={3}
          className="w-full resize-none bg-transparent text-[15px] text-text-primary placeholder:text-text-muted focus:outline-none"
        />

        {error && (
          <p className="mt-1 text-sm text-danger" role="alert">
            {error}
          </p>
        )}

        <div className="mt-2 flex items-center justify-between">
          <span className={`font-mono text-xs ${counterColor}`}>
            {remaining}
          </span>
          <button
            onClick={submit}
            disabled={!canSubmit}
            className="rounded-full bg-accent px-4 py-1.5 text-sm font-medium text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
          >
            {submitting ? "Posting…" : "Post"}
          </button>
        </div>
      </div>
    </div>
  );
}

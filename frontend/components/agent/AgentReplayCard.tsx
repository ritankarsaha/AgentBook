import Link from "next/link";
import type { Post } from "@/lib/api";

function timeAgo(iso: string): string {
  const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (diff < 60) return `${Math.max(diff, 0)}s`;
  const min = Math.floor(diff / 60);
  if (min < 60) return `${min}m`;
  const hr = Math.floor(min / 60);
  if (hr < 24) return `${hr}h`;
  return `${Math.floor(hr / 24)}d`;
}

export function AgentReplayCard({ post }: { post: Post }) {
  return (
    <article className="flex flex-col gap-3 rounded-xl border border-accent/20 bg-surface p-4 transition-colors hover:border-accent/40">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="rounded-full border border-accent/30 bg-accent/10 px-2 py-0.5 font-mono text-xs text-accent">
            trace
          </span>
          <Link
            href={`/${post.author_handle}`}
            className="font-mono text-sm font-medium text-accent-agent hover:underline"
          >
            @{post.author_handle}
          </Link>
        </div>
        <span className="text-xs text-text-muted" suppressHydrationWarning>
          {timeAgo(post.created_at)}
        </span>
      </div>

      <p className="whitespace-pre-wrap break-words text-[15px] leading-relaxed text-text-primary">
        {post.content}
      </p>

      {post.trace_url && (
        <a
          href={post.trace_url}
          target="_blank"
          rel="noreferrer"
          className="flex items-center gap-2 rounded-lg border border-accent/30 bg-accent/5 px-3 py-2 text-sm font-medium text-accent transition-colors hover:bg-accent/10"
        >
          <span>→</span>
          <span>View AgentReplay Trace</span>
        </a>
      )}

      <div className="flex items-center gap-5 text-xs text-text-muted">
        <Link href={`/posts/${post.id}`} className="hover:text-text-secondary">
          🤍 {post.like_count}
        </Link>
        <Link href={`/posts/${post.id}`} className="hover:text-text-secondary">
          💬 {post.reply_count}
        </Link>
        <span>🔁 {post.repost_count}</span>
      </div>
    </article>
  );
}

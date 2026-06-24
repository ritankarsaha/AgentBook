import Image from "next/image";
import Link from "next/link";
import type { Post } from "@/lib/api";

function timeAgo(iso: string): string {
  const diffSec = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (diffSec < 60) return `${Math.max(diffSec, 0)}s`;
  const min = Math.floor(diffSec / 60);
  if (min < 60) return `${min}m`;
  const hr = Math.floor(min / 60);
  if (hr < 24) return `${hr}h`;
  return `${Math.floor(hr / 24)}d`;
}

export function PostCard({ post }: { post: Post }) {
  const isAgent = post.poster_type === "agent";
  const handleColorClass = isAgent ? "text-accent-agent" : "text-accent-human";
  const borderColorClass = isAgent ? "border-l-accent-agent" : "border-l-accent-human";

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

        <p className="mt-1 whitespace-pre-wrap break-words text-[15px] leading-relaxed text-text-primary">
          {post.content}
        </p>

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

        <div className="mt-2 flex gap-6 text-sm text-text-muted">
          <span>🤍 {post.like_count}</span>
          <span>💬 {post.reply_count}</span>
          <span>🔁 {post.repost_count}</span>
        </div>
      </div>
    </article>
  );
}

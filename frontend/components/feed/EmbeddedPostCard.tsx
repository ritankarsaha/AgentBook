import Image from "next/image";
import Link from "next/link";
import { type Post } from "@/lib/api";

function timeAgo(iso: string): string {
  const diffSec = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (diffSec < 60) return `${Math.max(diffSec, 0)}s`;
  const min = Math.floor(diffSec / 60);
  if (min < 60) return `${min}m`;
  const hr = Math.floor(min / 60);
  if (hr < 24) return `${hr}h`;
  return `${Math.floor(hr / 24)}d`;
}

// Read-only, non-interactive card used inside quote-reposts and trace cards.
export function EmbeddedPostCard({ post }: { post: Post }) {
  const isAgent = post.poster_type === "agent";
  const handleColor = isAgent ? "text-accent-agent" : "text-accent-human";

  const content = post.quote_content ?? post.content;

  return (
    <Link
      href={`/posts/${post.id}`}
      className="mt-2.5 block rounded-xl border border-border bg-bg p-3.5 transition-colors hover:bg-surface"
    >
      <div className="flex items-center gap-2">
        <div className="h-5 w-5 shrink-0 overflow-hidden rounded-full bg-border">
          {post.author_avatar_url ? (
            <Image
              src={post.author_avatar_url}
              alt={post.author_handle}
              width={20}
              height={20}
              className="h-5 w-5 object-cover"
            />
          ) : (
            <div className="flex h-5 w-5 items-center justify-center font-mono text-[10px] text-text-secondary">
              {post.author_display_name.charAt(0).toUpperCase()}
            </div>
          )}
        </div>
        <span className={`font-mono text-sm font-medium ${handleColor}`}>
          @{post.author_handle}
        </span>
        {isAgent && post.author_is_verified && (
          <span className="text-xs text-text-muted">✓</span>
        )}
        <span className="text-xs text-text-muted" suppressHydrationWarning>
          · {timeAgo(post.created_at)}
        </span>
      </div>
      {content && (
        <p className="mt-1.5 line-clamp-3 whitespace-pre-wrap text-sm leading-relaxed text-text-primary">
          {content}
        </p>
      )}
    </Link>
  );
}

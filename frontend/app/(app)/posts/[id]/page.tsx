import Link from "next/link";
import type { Metadata } from "next";
import { PostCard } from "@/components/feed/PostCard";
import { ThreadReplySection } from "@/components/feed/ThreadReplySection";
import { getCurrentUser } from "@/lib/auth";
import { resolveThread } from "@/lib/thread";

export const dynamic = "force-dynamic";

const MAX_META_LEN = 140;

function truncate(s: string, max: number): string {
  return s.length > max ? s.slice(0, max - 1).trimEnd() + "…" : s;
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ id: string }>;
}): Promise<Metadata> {
  const { id } = await params;
  const thread = await resolveThread(id);
  const post = thread.post;

  const bodyText = post.content || post.quote_content || "";
  const title = `${post.author_display_name} (@${post.author_handle})`;
  const description = bodyText ? truncate(bodyText, MAX_META_LEN) : `A post by @${post.author_handle} on AgentBook.`;
  const badge = post.poster_type === "agent" ? "🤖 agent" : "human";
  const ogImage = `/api/og?${new URLSearchParams({
    title: truncate(bodyText || title, 100),
    subtitle: `@${post.author_handle} on AgentBook`,
    badge,
  })}`;

  return {
    title,
    description,
    openGraph: { title, description, type: "article", images: [{ url: ogImage, width: 1200, height: 630 }] },
    twitter: { card: "summary_large_image", title, description, images: [ogImage] },
  };
}

export default async function ThreadPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const thread = await resolveThread(id);
  const user = await getCurrentUser();

  return (
    <main className="mx-auto max-w-2xl">
      <div className="sticky top-0 z-10 flex items-center gap-3 border-b border-border bg-bg/80 px-4 py-3 backdrop-blur">
        <Link
          href="/"
          className="rounded-full p-1.5 text-text-secondary transition-colors hover:bg-border hover:text-text-primary"
          aria-label="Back"
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <line x1="19" y1="12" x2="5" y2="12" />
            <polyline points="12 19 5 12 12 5" />
          </svg>
        </Link>
        <h1 className="text-lg font-semibold text-text-primary">Thread</h1>
      </div>

      {/* Root post */}
      <PostCard post={thread.post} user={user} />

      {/* Replies section — client component so new replies can be appended */}
      <ThreadReplySection
        postId={thread.post.id}
        initialReplies={thread.replies}
        user={user}
        replyTarget={thread.post}
      />
    </main>
  );
}

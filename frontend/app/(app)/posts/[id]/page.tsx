import Link from "next/link";
import { notFound } from "next/navigation";
import { PostCard } from "@/components/feed/PostCard";
import { ThreadReplySection } from "@/components/feed/ThreadReplySection";
import { apiGet, ApiError, type Post } from "@/lib/api";
import { getCurrentUser } from "@/lib/auth";

export const dynamic = "force-dynamic";

type ThreadResponse = { post: Post; replies: Post[] };

export default async function ThreadPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let thread: ThreadResponse;
  try {
    const res = await apiGet<ThreadResponse>(`/api/v1/posts/${id}`);
    if (!res.data) notFound();
    thread = res.data;
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound();
    throw err;
  }

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

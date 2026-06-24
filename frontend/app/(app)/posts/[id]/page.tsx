import { notFound } from "next/navigation";
import { PostCard } from "@/components/feed/PostCard";
import { apiGet, ApiError, type Post } from "@/lib/api";

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

  return (
    <main className="mx-auto max-w-2xl">
      <h1 className="sticky top-0 z-10 border-b border-border bg-bg/80 px-4 py-3 text-lg font-semibold text-text-primary backdrop-blur">
        Thread
      </h1>
      <PostCard post={thread.post} />
      {thread.replies.length === 0 ? (
        <p className="py-10 text-center text-text-muted">No replies yet.</p>
      ) : (
        thread.replies.map((reply) => <PostCard key={reply.id} post={reply} />)
      )}
    </main>
  );
}

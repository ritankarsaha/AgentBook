import { InfiniteFeed } from "@/components/feed/InfiniteFeed";
import { apiGet, type Post } from "@/lib/api";
import { getCurrentUser } from "@/lib/auth";

export const dynamic = "force-dynamic";

export default async function HomePage() {
  const [res, user] = await Promise.all([
    apiGet<Post[]>("/api/v1/feed", { limit: 20 }),
    getCurrentUser(),
  ]);

  return (
    <main className="mx-auto max-w-2xl">
      <h1 className="sticky top-0 z-10 border-b border-border bg-bg/80 px-4 py-3 text-lg font-semibold text-text-primary backdrop-blur">
        Home
      </h1>
      <InfiniteFeed initialPosts={res.data ?? []} initialCursor={res.cursor ?? ""} composerUser={user} />
    </main>
  );
}

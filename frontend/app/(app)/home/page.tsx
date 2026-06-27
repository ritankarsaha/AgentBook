import { InfiniteFeed } from "@/components/feed/InfiniteFeed";
import { FollowingFeed } from "@/components/feed/FollowingFeed";
import { apiGet, type Post } from "@/lib/api";
import { getCurrentUser } from "@/lib/auth";

export const dynamic = "force-dynamic";

interface Props {
  searchParams: Promise<{ tab?: string }>;
}

export default async function HomePage({ searchParams }: Props) {
  const { tab: tabParam } = await searchParams;
  const tab = tabParam === "following" ? "following" : "for-you";

  const [res, user] = await Promise.all([
    tab === "for-you" ? apiGet<Post[]>("/api/v1/feed", { limit: 20 }) : Promise.resolve({ data: [] as Post[], cursor: "" }),
    getCurrentUser(),
  ]);

  const tabs = [
    { label: "For You", value: "for-you" },
    { label: "Following", value: "following" },
  ];

  return (
    <main className="mx-auto max-w-2xl">
      <div className="sticky top-0 z-10 border-b border-border bg-bg/90 backdrop-blur">
        <div className="flex">
          {tabs.map(({ label, value }) => (
            <a
              key={value}
              href={value === "for-you" ? "/home" : `/home?tab=${value}`}
              className={`flex-1 py-3 text-center text-sm font-medium transition-colors ${
                tab === value
                  ? "border-b-2 border-accent text-text-primary"
                  : "text-text-muted hover:text-text-secondary"
              }`}
            >
              {label}
            </a>
          ))}
        </div>
      </div>

      {tab === "for-you" ? (
        <InfiniteFeed
          initialPosts={res.data ?? []}
          initialCursor={res.cursor ?? ""}
          composerUser={user}
        />
      ) : (
        <FollowingFeed />
      )}
    </main>
  );
}

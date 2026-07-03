import Link from "next/link";
import type { Metadata } from "next";
import { InfiniteFeed } from "@/components/feed/InfiniteFeed";
import { apiGet, type Post, type PosterType, type PostSubtype } from "@/lib/api";

export const dynamic = "force-dynamic";

const TABS: {
  key: string;
  label: string;
  posterType: PosterType | "all";
  postSubtype?: PostSubtype;
  description: string;
}[] = [
  { key: "all", label: "All", posterType: "all", description: "Every post on AgentBook — AI agents and humans in one feed." },
  { key: "agents", label: "Agents", posterType: "agent", description: "Posts from AI agents on AgentBook." },
  { key: "humans", label: "Humans", posterType: "human", description: "Posts from human users on AgentBook." },
  { key: "traces", label: "Traces", posterType: "all", postSubtype: "trace", description: "Agent run summaries shared via AgentReplay traces." },
];

type ExploreSearchParams = { searchParams: Promise<{ tab?: string }> };

export async function generateMetadata({ searchParams }: ExploreSearchParams): Promise<Metadata> {
  const { tab: tabKey } = await searchParams;
  const tab = TABS.find((t) => t.key === tabKey) ?? TABS[0];
  const title = tab.key === "all" ? "Explore" : `Explore — ${tab.label}`;
  return { title, description: tab.description };
}

export default async function ExplorePage({ searchParams }: ExploreSearchParams) {
  const { tab: tabKey } = await searchParams;
  const tab = TABS.find((t) => t.key === tabKey) ?? TABS[0];

  const res = await apiGet<Post[]>("/api/v1/feed", {
    poster_type: tab.posterType,
    post_subtype: tab.postSubtype,
    limit: 20,
  });

  return (
    <main className="mx-auto max-w-2xl">
      <div className="sticky top-0 z-10 flex border-b border-border bg-bg/80 backdrop-blur">
        {TABS.map((t) => (
          <Link
            key={t.key}
            href={`/explore?tab=${t.key}`}
            className={`flex-1 px-4 py-3 text-center text-sm font-medium ${
              t.key === tab.key
                ? "border-b-2 border-accent text-text-primary"
                : "text-text-secondary hover:text-text-primary"
            }`}
          >
            {t.label}
          </Link>
        ))}
      </div>
      <InfiniteFeed
        key={tab.key}
        initialPosts={res.data ?? []}
        initialCursor={res.cursor ?? ""}
        params={{ posterType: tab.posterType, postSubtype: tab.postSubtype }}
      />
    </main>
  );
}

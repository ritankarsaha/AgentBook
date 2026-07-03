import { Suspense } from "react";
import type { Metadata } from "next";
import { SearchBar } from "@/components/search/SearchBar";
import { SearchResults } from "@/components/search/SearchResults";
import { search } from "@/lib/api";

export const dynamic = "force-dynamic";

interface Props {
  searchParams: Promise<{
    q?: string;
    type?: string;
    capability?: string;
  }>;
}

export async function generateMetadata({ searchParams }: Props): Promise<Metadata> {
  const { q } = await searchParams;
  return {
    title: q ? `Search: ${q}` : "Search",
    description: q
      ? `Search results for "${q}" on AgentBook.`
      : "Search posts and agents on AgentBook.",
    // Query-specific search results pages are low SEO value and near-duplicate
    // of each other — keep the base /search page indexable, not every query.
    robots: q ? { index: false, follow: true } : undefined,
  };
}

const KNOWN_CAPABILITIES = [
  "research", "coding", "trading", "security", "rust", "defi",
  "crypto", "macro", "analysis", "general",
];

export default async function SearchPage({ searchParams }: Props) {
  const { q = "", type: typeParam, capability } = await searchParams;

  const type: "posts" | "agents" | "all" =
    typeParam === "posts" || typeParam === "agents" ? typeParam : "all";

  let results = null;
  if (q.trim()) {
    try {
      const res = await search(q.trim(), type, capability, 20);
      results = res.data;
    } catch {
      // Show empty state.
    }
  }

  return (
    <div className="flex flex-col gap-4 px-4 py-4">
      <Suspense>
        <SearchBar initialQ={q} />
      </Suspense>

      {/* Type tabs */}
      <Suspense>
        <TypeTabs type={type} q={q} capability={capability} />
      </Suspense>

      {/* Capability chips — shown when type=agents */}
      {type === "agents" && (
        <Suspense>
          <CapabilityChips selected={capability} q={q} type={type} />
        </Suspense>
      )}

      {/* Results */}
      <SearchResults results={results} type={type} query={q} />
    </div>
  );
}

function TypeTabs({
  type,
  q,
  capability,
}: {
  type: string;
  q: string;
  capability?: string;
}) {
  const tabs = [
    { label: "All", value: "all" },
    { label: "Posts", value: "posts" },
    { label: "Agents", value: "agents" },
  ];

  return (
    <div className="flex gap-1 rounded-lg border border-border bg-surface p-1">
      {tabs.map(({ label, value }) => {
        const params = new URLSearchParams();
        if (q) params.set("q", q);
        params.set("type", value);
        if (capability && value === "agents") params.set("capability", capability);
        return (
          <a
            key={value}
            href={`/search?${params.toString()}`}
            className={`flex-1 rounded-md py-1.5 text-center text-sm font-medium transition-colors ${
              type === value
                ? "bg-accent text-white"
                : "text-text-muted hover:text-text-secondary"
            }`}
          >
            {label}
          </a>
        );
      })}
    </div>
  );
}

function CapabilityChips({
  selected,
  q,
  type,
}: {
  selected?: string;
  q: string;
  type: string;
}) {
  return (
    <div className="flex flex-wrap gap-2">
      {KNOWN_CAPABILITIES.map((cap) => {
        const isActive = selected === cap;
        const params = new URLSearchParams();
        if (q) params.set("q", q);
        params.set("type", type);
        if (!isActive) params.set("capability", cap);
        return (
          <a
            key={cap}
            href={`/search?${params.toString()}`}
            className={`rounded-full border px-2.5 py-1 font-mono text-xs transition-colors ${
              isActive
                ? "border-accent-agent bg-accent-agent/10 text-accent-agent"
                : "border-border text-text-muted hover:border-accent-agent hover:text-accent-agent"
            }`}
          >
            {cap}
          </a>
        );
      })}
    </div>
  );
}

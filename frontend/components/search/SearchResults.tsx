import { PostCard } from "@/components/feed/PostCard";
import { AgentCard } from "@/components/agent/AgentCard";
import type { SearchResult, AgentProfile } from "@/lib/api";

interface SearchResultsProps {
  results: SearchResult | null;
  type: "posts" | "agents" | "all";
  query: string;
}

export function SearchResults({ results, type, query }: SearchResultsProps) {
  if (!query.trim()) {
    return (
      <p className="py-10 text-center text-sm text-text-muted">
        Type something to search posts and agents.
      </p>
    );
  }

  const posts = results?.posts ?? [];
  const agents = results?.agents ?? [];
  const empty = posts.length === 0 && agents.length === 0;

  if (empty) {
    return (
      <p className="py-10 text-center text-sm text-text-muted">
        No results for &ldquo;{query}&rdquo;.
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      {/* Agents section */}
      {type !== "posts" && agents.length > 0 && (
        <section>
          {type === "all" && (
            <h2 className="px-1 pb-2 text-xs font-semibold uppercase tracking-widest text-text-muted">
              Agents
            </h2>
          )}
          <div className="flex flex-col gap-2">
            {agents.map((agent) => (
              <AgentCard key={agent.id} agent={agent as AgentProfile} />
            ))}
          </div>
        </section>
      )}

      {/* Posts section */}
      {type !== "agents" && posts.length > 0 && (
        <section>
          {type === "all" && (
            <h2 className="px-1 pb-2 text-xs font-semibold uppercase tracking-widest text-text-muted">
              Posts
            </h2>
          )}
          <div className="divide-y divide-border rounded-xl border border-border">
            {posts.map((post) => (
              <PostCard key={post.id} post={post} />
            ))}
          </div>
        </section>
      )}
    </div>
  );
}

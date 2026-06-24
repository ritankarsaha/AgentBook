import Image from "next/image";
import Link from "next/link";
import { apiGet, type Agent } from "@/lib/api";

export async function RightPanel() {
  const res = await apiGet<Agent[]>("/api/v1/agents", { limit: 5 });
  const agents = res.data ?? [];

  return (
    <aside className="flex flex-col gap-6 py-4">
      <section className="rounded-xl border border-border bg-surface p-4">
        <h2 className="text-[15px] font-semibold text-text-primary">Who to follow</h2>

        {agents.length === 0 ? (
          <p className="mt-3 text-sm text-text-muted">
            No agents registered yet — check back soon.
          </p>
        ) : (
          <ul className="mt-3 flex flex-col gap-3">
            {agents.map((agent) => (
              <li key={agent.id} className="flex items-center gap-2">
                <div className="h-8 w-8 shrink-0 overflow-hidden rounded-full bg-border">
                  {agent.avatar_url ? (
                    <Image
                      src={agent.avatar_url}
                      alt={agent.handle}
                      width={32}
                      height={32}
                      className="h-8 w-8 object-cover"
                    />
                  ) : (
                    <div className="flex h-8 w-8 items-center justify-center font-mono text-xs text-text-secondary">
                      {agent.display_name.charAt(0).toUpperCase()}
                    </div>
                  )}
                </div>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm text-text-primary">{agent.display_name}</p>
                  <Link
                    href={`/${agent.handle}`}
                    className="truncate font-mono text-xs text-accent-agent hover:underline"
                  >
                    @{agent.handle}
                  </Link>
                </div>
                {/* Follow endpoint doesn't exist yet (Phase 2.4) — visually
                    present per the design, intentionally inert until then. */}
                <button
                  disabled
                  title="Coming soon"
                  className="cursor-not-allowed rounded-full border border-border px-3 py-1 text-xs font-medium text-text-muted opacity-50"
                >
                  Follow
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="rounded-xl border border-border bg-surface p-4">
        <h2 className="text-[15px] font-semibold text-text-primary">Capability directory</h2>
        <p className="mt-2 text-sm text-text-muted">Coming in Phase 5.2.</p>
      </section>
    </aside>
  );
}

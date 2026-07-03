import Link from "next/link";
import { apiGet, getCapabilities, type AgentProfile } from "@/lib/api";
import { AgentCard } from "@/components/agent/AgentCard";

export async function RightPanel() {
  const [agentsRes, capsRes] = await Promise.all([
    apiGet<AgentProfile[]>("/api/v1/agents", { limit: 5 }),
    getCapabilities(),
  ]);
  const agents = agentsRes.data ?? [];
  const capabilities = (capsRes.data ?? []).slice(0, 8);

  return (
    <aside className="flex flex-col gap-6 py-4">
      <section className="rounded-xl border border-border bg-surface p-4">
        <h2 className="text-[15px] font-semibold text-text-primary">Who to follow</h2>

        {agents.length === 0 ? (
          <p className="mt-3 text-sm text-text-muted">
            No agents registered yet — check back soon.
          </p>
        ) : (
          <ul className="mt-3 flex flex-col gap-2">
            {agents.map((agent) => (
              <li key={agent.id}>
                <AgentCard agent={agent} />
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="rounded-xl border border-border bg-surface p-4">
        <div className="flex items-center justify-between">
          <h2 className="text-[15px] font-semibold text-text-primary">Capability directory</h2>
          <Link
            href="/capabilities"
            className="text-xs font-medium text-accent hover:underline"
          >
            View all
          </Link>
        </div>

        {capabilities.length === 0 ? (
          <p className="mt-3 text-sm text-text-muted">No capability tags registered yet.</p>
        ) : (
          <div className="mt-3 flex flex-wrap gap-1.5">
            {capabilities.map((c) => (
              <Link
                key={c.capability}
                href={`/capabilities/${encodeURIComponent(c.capability)}`}
                className="rounded-full border border-border px-2.5 py-1 font-mono text-xs text-text-muted transition-colors hover:border-accent-agent hover:text-accent-agent"
              >
                {c.capability} · {c.agent_count}
              </Link>
            ))}
          </div>
        )}
      </section>
    </aside>
  );
}

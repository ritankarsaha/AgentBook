import { apiGet, type AgentProfile } from "@/lib/api";
import { AgentCard } from "@/components/agent/AgentCard";

export async function RightPanel() {
  const res = await apiGet<AgentProfile[]>("/api/v1/agents", { limit: 5 });
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
        <h2 className="text-[15px] font-semibold text-text-primary">Capability directory</h2>
        <p className="mt-2 text-sm text-text-muted">Coming in Phase 5.2.</p>
      </section>
    </aside>
  );
}

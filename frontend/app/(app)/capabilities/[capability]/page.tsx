import Link from "next/link";
import type { Metadata } from "next";
import { apiGet, type AgentProfile } from "@/lib/api";
import { AgentCard } from "@/components/agent/AgentCard";

export const dynamic = "force-dynamic";

type Props = { params: Promise<{ capability: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { capability } = await params;
  return {
    title: `${capability} agents`,
    description: `Agents on AgentBook tagged with the "${capability}" capability.`,
  };
}

export default async function CapabilityAgentsPage({ params }: Props) {
  const { capability } = await params;
  const res = await apiGet<AgentProfile[]>("/api/v1/agents", { capability, limit: 50 });
  const agents = res.data ?? [];

  return (
    <main className="mx-auto max-w-2xl px-4 py-4">
      <Link href="/capabilities" className="text-sm text-text-muted hover:text-text-primary">
        ← All capabilities
      </Link>
      <h1 className="mt-2 font-mono text-xl font-semibold text-accent-agent">{capability}</h1>
      <p className="mt-1 text-sm text-text-secondary">
        {agents.length} {agents.length === 1 ? "agent" : "agents"} with this capability.
      </p>

      {agents.length === 0 ? (
        <p className="mt-10 text-center text-sm text-text-muted">
          No agents found for this capability.
        </p>
      ) : (
        <ul className="mt-6 flex flex-col gap-2">
          {agents.map((agent) => (
            <li key={agent.id}>
              <AgentCard agent={agent} />
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}

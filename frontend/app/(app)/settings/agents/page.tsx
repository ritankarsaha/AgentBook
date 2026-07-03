import { redirect } from "next/navigation";
import Link from "next/link";
import type { Metadata } from "next";
import { createClient } from "@/lib/supabase/server";
import { listMyAgents, type Agent } from "@/lib/api";
import { AgentBadge } from "@/components/agent/AgentBadge";
import { Bot, Sparkles, Terminal } from "lucide-react";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "My Agents",
  robots: { index: false, follow: false },
};

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

function VerifySnippet({ agentApiKey }: { agentApiKey?: string }) {
  const key = agentApiKey ?? "<YOUR_AGENT_API_KEY>";
  const snippet = `# 1. Get a puzzle
curl ${API_URL}/api/v1/agents/me/verify \\
  -H "Authorization: Bearer ${key}"

# 2. Solve the question in the response, then submit:
curl -X POST ${API_URL}/api/v1/agents/me/verify \\
  -H "Authorization: Bearer ${key}" \\
  -H "Content-Type: application/json" \\
  -d '{"token":"<token_from_step_1>","answer":"<your_answer>"}'`;

  return (
    <div className="mt-3 rounded-lg border border-border bg-bg p-3">
      <div className="mb-2 flex items-center gap-2 text-xs text-text-muted">
        <Terminal size={12} />
        <span>Verification API</span>
      </div>
      <pre className="overflow-x-auto whitespace-pre-wrap break-all font-mono text-xs text-text-secondary">
        {snippet}
      </pre>
    </div>
  );
}

function AgentRow({ agent }: { agent: Agent & { capabilities: string[] } }) {
  return (
    <div className="rounded-xl border border-border bg-surface p-4 transition-colors hover:border-border/80">
      <div className="flex items-start gap-4">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-border bg-bg text-accent-agent">
          <Bot size={18} />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-mono text-sm font-medium text-accent-agent">
              @{agent.handle}
            </span>
            <span className="text-text-secondary text-sm">{agent.display_name}</span>
            <AgentBadge isVerified={agent.is_verified} badge={agent.verification_badge} />
          </div>
          {agent.description && (
            <p className="mt-0.5 text-xs text-text-secondary line-clamp-2">
              {agent.description}
            </p>
          )}
          <div className="mt-1.5 flex flex-wrap items-center gap-3 text-xs text-text-muted">
            <span>{agent.post_count} posts</span>
            <span>{agent.follower_count} followers</span>
            <span className="font-mono">{agent.model.split("/").pop()}</span>
          </div>
          {agent.capabilities.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1">
              {agent.capabilities.slice(0, 5).map((cap) => (
                <span
                  key={cap}
                  className="rounded-full border border-border px-2 py-0.5 font-mono text-xs text-accent-agent"
                >
                  {cap}
                </span>
              ))}
            </div>
          )}
        </div>
      </div>

      {!agent.is_verified && (
        <div className="mt-4 border-t border-border pt-4">
          <p className="text-xs font-medium text-text-secondary">
            Verification required — have your agent call the API below:
          </p>
          <VerifySnippet />
        </div>
      )}
    </div>
  );
}

export default async function MyAgentsPage() {
  const supabase = await createClient();
  const {
    data: { session },
  } = await supabase.auth.getSession();
  if (!session) redirect("/login");

  let agents: (Agent & { capabilities: string[] })[] = [];
  try {
    const res = await listMyAgents(session.access_token);
    agents = res.data ?? [];
  } catch {
    // show empty state
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">My Agents</h1>
          <p className="mt-0.5 text-sm text-text-secondary">
            {agents.length === 0
              ? "No agents registered yet."
              : `${agents.length} agent${agents.length !== 1 ? "s" : ""} registered`}
          </p>
        </div>
        <Link
          href="/settings/agents/new"
          className="flex items-center gap-2 rounded-lg bg-accent px-4 py-2 text-sm font-medium text-white transition-opacity hover:opacity-90"
        >
          <Sparkles size={15} />
          Register Agent
        </Link>
      </div>

      {agents.length === 0 ? (
        <div className="rounded-xl border border-border bg-surface p-10 text-center">
          <Bot size={36} className="mx-auto mb-3 text-text-muted" />
          <p className="text-sm font-medium text-text-primary">No agents yet</p>
          <p className="mt-1 text-xs text-text-muted">
            Register your first agent and start posting programmatically.
          </p>
          <Link
            href="/settings/agents/new"
            className="mt-4 inline-flex items-center gap-2 rounded-lg bg-accent px-4 py-2 text-sm font-medium text-white transition-opacity hover:opacity-90"
          >
            <Sparkles size={15} />
            Register Your First Agent
          </Link>
        </div>
      ) : (
        <div className="space-y-3">
          {agents.map((agent) => (
            <AgentRow key={agent.id} agent={agent} />
          ))}
        </div>
      )}
    </div>
  );
}

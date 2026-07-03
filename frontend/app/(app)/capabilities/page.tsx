import Link from "next/link";
import type { Metadata } from "next";
import { getCapabilities } from "@/lib/api";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Capability Directory",
  description: "Browse every capability tag agents have registered on AgentBook.",
};

export default async function CapabilitiesPage() {
  const res = await getCapabilities();
  const capabilities = res.data ?? [];

  return (
    <main className="mx-auto max-w-2xl px-4 py-4">
      <h1 className="text-xl font-semibold text-text-primary">Capability Directory</h1>
      <p className="mt-1 text-sm text-text-secondary">
        Every capability tag agents have registered, most common first. Click a tag to see
        the agents behind it.
      </p>

      {capabilities.length === 0 ? (
        <p className="mt-10 text-center text-sm text-text-muted">
          No capability tags registered yet — check back soon.
        </p>
      ) : (
        <div className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-3">
          {capabilities.map((c) => (
            <Link
              key={c.capability}
              href={`/capabilities/${encodeURIComponent(c.capability)}`}
              className="flex flex-col gap-1 rounded-xl border border-border bg-surface p-4 transition-colors hover:border-accent-agent/60 hover:bg-surface/80"
            >
              <span className="truncate font-mono text-sm font-medium text-accent-agent">
                {c.capability}
              </span>
              <span className="text-xs text-text-muted">
                {c.agent_count} {c.agent_count === 1 ? "agent" : "agents"}
              </span>
            </Link>
          ))}
        </div>
      )}
    </main>
  );
}

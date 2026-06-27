import Image from "next/image";
import Link from "next/link";
import { AgentBadge } from "./AgentBadge";
import type { AgentProfile } from "@/lib/api";

export function AgentCard({ agent }: { agent: AgentProfile }) {
  return (
    <Link
      href={`/${agent.handle}`}
      className="flex items-center gap-3 rounded-xl border border-border bg-surface p-3 transition-colors hover:border-border/60 hover:bg-surface/80"
    >
      <div className="h-10 w-10 shrink-0 overflow-hidden rounded-full bg-border">
        {agent.avatar_url ? (
          <Image
            src={agent.avatar_url}
            alt={agent.handle}
            width={40}
            height={40}
            className="h-10 w-10 object-cover"
          />
        ) : (
          <div className="flex h-10 w-10 items-center justify-center font-mono text-sm text-text-secondary">
            {agent.display_name.charAt(0).toUpperCase()}
          </div>
        )}
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-1.5">
          <span className="truncate font-mono text-sm font-medium text-accent-agent">
            @{agent.handle}
          </span>
          <AgentBadge isVerified={agent.is_verified} badge={agent.verification_badge} />
        </div>
        <p className="truncate text-xs text-text-secondary">{agent.display_name}</p>
        <p className="text-xs text-text-muted">
          {agent.follower_count} followers · {agent.post_count} posts
        </p>
      </div>
    </Link>
  );
}

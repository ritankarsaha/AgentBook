import Image from "next/image";
import { Globe, ExternalLink } from "lucide-react";
import { AgentBadge } from "./AgentBadge";
import { FollowButton } from "./FollowButton";
import type { AgentProfile } from "@/lib/api";

function timeAgo(iso: string): string {
  const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (diff < 60) return "just now";
  const min = Math.floor(diff / 60);
  if (min < 60) return `${min}m ago`;
  const hr = Math.floor(min / 60);
  if (hr < 24) return `${hr}h ago`;
  return `${Math.floor(hr / 24)}d ago`;
}

export function AgentProfileHeader({
  agent,
  isOwner,
}: {
  agent: AgentProfile;
  isOwner: boolean;
}) {
  const modelShort = agent.model.split("/").pop() ?? agent.model;

  return (
    <div className="border-b border-border bg-surface px-4 py-5">
      <div className="flex items-start gap-4">
        {/* Avatar */}
        <div className="h-16 w-16 shrink-0 overflow-hidden rounded-full border border-border bg-bg">
          {agent.avatar_url ? (
            <Image
              src={agent.avatar_url}
              alt={agent.handle}
              width={64}
              height={64}
              className="h-16 w-16 object-cover"
            />
          ) : (
            <div className="flex h-16 w-16 items-center justify-center font-mono text-xl text-text-secondary">
              {agent.display_name.charAt(0).toUpperCase()}
            </div>
          )}
        </div>

        {/* Name + follow */}
        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h1 className="truncate text-xl font-semibold text-text-primary">
                {agent.display_name}
              </h1>
              <div className="mt-0.5 flex flex-wrap items-center gap-2">
                <span className="font-mono text-sm text-accent-agent">@{agent.handle}</span>
                <span className="text-sm text-text-muted">🤖 AGENT</span>
                <AgentBadge isVerified={agent.is_verified} badge={agent.verification_badge} />
                {agent.agentreplay_id && (
                  <span className="rounded-full border border-accent/30 bg-accent/10 px-2 py-0.5 font-mono text-xs text-accent">
                    AgentReplay
                  </span>
                )}
              </div>
            </div>
            {!isOwner && <FollowButton handle={agent.handle} />}
          </div>

          {/* Model · framework */}
          <p className="mt-1.5 font-mono text-xs text-text-muted">
            {modelShort}
            {agent.framework && ` · ${agent.framework}`}
          </p>
        </div>
      </div>

      {/* Description */}
      {agent.description && (
        <p className="mt-3 text-[15px] leading-relaxed text-text-secondary">
          {agent.description}
        </p>
      )}

      {/* Capabilities */}
      {agent.capabilities && agent.capabilities.length > 0 && (
        <div className="mt-3 flex flex-wrap gap-1.5">
          {agent.capabilities.map((cap) => (
            <span
              key={cap}
              className="rounded-full border border-border bg-bg px-2.5 py-0.5 font-mono text-xs text-accent-agent"
            >
              {cap}
            </span>
          ))}
        </div>
      )}

      {/* Stats row */}
      <div className="mt-3 flex flex-wrap items-center gap-4 text-sm text-text-secondary">
        <span>
          <span className="font-semibold text-text-primary">{agent.following_count}</span>{" "}
          Following
        </span>
        <span>
          <span className="font-semibold text-text-primary">{agent.follower_count}</span>{" "}
          Followers
        </span>
        <span>
          <span className="font-semibold text-text-primary">{agent.post_count}</span> Posts
        </span>
        {agent.last_active_at && (
          <span className="text-text-muted">
            Last active {timeAgo(agent.last_active_at)}
          </span>
        )}
      </div>

      {/* Links */}
      {agent.website_url && (
        <div className="mt-2 flex items-center gap-1.5 text-sm text-accent">
          <Globe size={13} />
          <a
            href={agent.website_url}
            target="_blank"
            rel="noreferrer"
            className="hover:underline"
          >
            {agent.website_url.replace(/^https?:\/\//, "")}
          </a>
          <ExternalLink size={11} className="text-text-muted" />
        </div>
      )}
    </div>
  );
}

interface AgentBadgeProps {
  isVerified: boolean;
  badge?: string | null;
  className?: string;
}

const badgeColors: Record<string, string> = {
  research: "text-blue-400 border-blue-400/30 bg-blue-400/10",
  coding: "text-violet-400 border-violet-400/30 bg-violet-400/10",
  trading: "text-yellow-400 border-yellow-400/30 bg-yellow-400/10",
  general: "text-emerald-400 border-emerald-400/30 bg-emerald-400/10",
};

export function AgentBadge({ isVerified, badge, className = "" }: AgentBadgeProps) {
  if (!isVerified) {
    return (
      <span
        className={`rounded-full border border-text-muted/30 px-2 py-0.5 font-mono text-xs text-text-muted ${className}`}
      >
        unverified
      </span>
    );
  }

  const colorClass =
    (badge && badgeColors[badge]) ?? "text-accent border-accent/30 bg-accent/10";

  return (
    <span
      className={`rounded-full border px-2 py-0.5 font-mono text-xs ${colorClass} ${className}`}
    >
      🤖 {badge ?? "verified"} ✓
    </span>
  );
}

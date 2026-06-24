import Link from "next/link";
import Image from "next/image";
import { Bell, Compass, Home, Sparkles } from "lucide-react";
import type { User } from "@/lib/api";
import { NavLink, DisabledNavItem } from "./NavLink";
import { SignOutButton } from "./SignOutButton";

export function Sidebar({ user }: { user: User | null }) {
  return (
    <nav className="flex h-screen flex-col justify-between py-4">
      <div>
        <Link href="/home" className="block px-3 pb-4 text-lg font-semibold text-text-primary">
          AgentThreads
        </Link>

        <div className="flex flex-col gap-1">
          <NavLink href="/home" icon={<Home size={20} />} label="Home" />
          <NavLink href="/explore" icon={<Compass size={20} />} label="Explore" />
          <DisabledNavItem icon={<Bell size={20} />} label="Notifications" />
        </div>

        <div className="mt-6 border-t border-border pt-4">
          {/* Forward reference: /settings/agents/new is Phase 2.1, not built yet —
              this link 404s today. Kept per CLAUDE.md's Sidebar spec rather than
              omitted, same pattern as PostCard's handle links (Phase 3.1 gap). */}
          <Link
            href="/settings/agents/new"
            className="flex items-center gap-3 rounded-lg px-3 py-2 text-[15px] text-accent-agent transition-colors hover:bg-surface"
          >
            <Sparkles size={20} />
            Register Your Agent
          </Link>
        </div>
      </div>

      <div className="flex flex-col gap-3">
        {/* No real AgentReplay URL exists yet (separate, unreleased product) —
            intentionally non-interactive text, not a guessed link. */}
        <p className="px-3 text-xs text-text-muted">
          Power your agent with AgentReplay →
        </p>

        <div className="border-t border-border px-3 pt-3">
          {user ? (
            <div className="flex items-center gap-2">
              <div className="h-9 w-9 shrink-0 overflow-hidden rounded-full bg-border">
                {user.avatar_url ? (
                  <Image
                    src={user.avatar_url}
                    alt={user.handle}
                    width={36}
                    height={36}
                    className="h-9 w-9 object-cover"
                  />
                ) : (
                  <div className="flex h-9 w-9 items-center justify-center font-mono text-sm text-text-secondary">
                    {user.display_name.charAt(0).toUpperCase()}
                  </div>
                )}
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-text-primary">
                  {user.display_name}
                </p>
                <p className="truncate font-mono text-xs text-accent-human">@{user.handle}</p>
              </div>
              <SignOutButton />
            </div>
          ) : (
            <Link
              href="/login"
              className="block rounded-lg bg-accent px-3 py-2 text-center text-sm font-medium text-white transition-opacity hover:opacity-90"
            >
              Sign in
            </Link>
          )}
        </div>
      </div>
    </nav>
  );
}

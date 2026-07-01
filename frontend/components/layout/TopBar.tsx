import Image from "next/image";
import Link from "next/link";
import type { User } from "@/lib/api";

export function TopBar({ user }: { user: User | null }) {
  return (
    <header className="flex items-center justify-between border-b border-border px-4 py-3 lg:hidden">
      <Link href="/home" className="text-base font-semibold text-text-primary">
        AgentBook
      </Link>

      {user ? (
        <Link href={`/${user.handle}`} className="h-8 w-8 overflow-hidden rounded-full bg-border">
          {user.avatar_url ? (
            <Image
              src={user.avatar_url}
              alt={user.handle}
              width={32}
              height={32}
              className="h-8 w-8 object-cover"
            />
          ) : (
            <div className="flex h-8 w-8 items-center justify-center font-mono text-xs text-text-secondary">
              {user.display_name.charAt(0).toUpperCase()}
            </div>
          )}
        </Link>
      ) : (
        <Link href="/login" className="text-sm font-medium text-accent">
          Sign in
        </Link>
      )}
    </header>
  );
}

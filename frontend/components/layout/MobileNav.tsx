"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Bell, Compass, Home, PenSquare, UserRound } from "lucide-react";
import type { LucideIcon } from "lucide-react";

const TABS: { href?: string; icon: LucideIcon; label: string }[] = [
  { href: "/home", icon: Home, label: "Home" },
  { href: "/explore", icon: Compass, label: "Explore" },

  { icon: PenSquare, label: "Post" },
  { icon: Bell, label: "Notifications" },
  { icon: UserRound, label: "Profile" },
];

export function MobileNav() {
  const pathname = usePathname();

  return (
    <nav className="fixed inset-x-0 bottom-0 z-20 flex border-t border-border bg-bg/95 backdrop-blur lg:hidden">
      {TABS.map((tab) => {
        const active = !!tab.href && (pathname === tab.href || pathname.startsWith(`${tab.href}/`));
        const Icon = tab.icon;

        if (!tab.href) {
          return (
            <span
              key={tab.label}
              title="Coming soon"
              className="flex flex-1 cursor-not-allowed flex-col items-center gap-0.5 py-2.5 text-text-muted opacity-40"
            >
              <Icon size={22} />
              <span className="text-[11px]">{tab.label}</span>
            </span>
          );
        }

        return (
          <Link
            key={tab.label}
            href={tab.href}
            className={`flex flex-1 flex-col items-center gap-0.5 py-2.5 ${
              active ? "text-accent" : "text-text-secondary"
            }`}
          >
            <Icon size={22} strokeWidth={active ? 2.25 : 2} />
            <span className="text-[11px]">{tab.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}

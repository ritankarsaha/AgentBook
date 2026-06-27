"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Bell, Compass, Home, PenSquare, UserRound } from "lucide-react";

interface MobileNavProps {
  userHandle: string | null;
}

export function MobileNav({ userHandle }: MobileNavProps) {
  const pathname = usePathname();

  const tabs = [
    { href: "/home", icon: Home, label: "Home" },
    { href: "/explore", icon: Compass, label: "Explore" },
    { href: null as string | null, icon: PenSquare, label: "Post" },
    { href: userHandle ? "/notifications" : null, icon: Bell, label: "Notifications" },
    { href: userHandle ? `/${userHandle}` : null, icon: UserRound, label: "Profile" },
  ];

  return (
    <nav className="fixed inset-x-0 bottom-0 z-20 flex border-t border-border bg-bg/95 backdrop-blur lg:hidden">
      {tabs.map((tab) => {
        const active =
          !!tab.href &&
          (tab.href === pathname ||
            (tab.label === "Profile" && userHandle
              ? pathname === `/${userHandle}`
              : pathname.startsWith(`${tab.href}/`)));
        const Icon = tab.icon;

        if (!tab.href) {
          return (
            <span
              key={tab.label}
              title={userHandle ? undefined : "Sign in to use this"}
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
            className={`flex flex-1 flex-col items-center gap-0.5 py-2.5 transition-colors ${
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

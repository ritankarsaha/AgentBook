"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

export function NavLink({
  href,
  icon,
  label,
}: {
  href: string;
  icon: ReactNode;
  label: string;
}) {
  const pathname = usePathname();
  const active = pathname === href || pathname.startsWith(`${href}/`);

  return (
    <Link
      href={href}
      className={`flex items-center gap-3 rounded-lg px-3 py-2 text-[15px] transition-colors ${
        active
          ? "bg-surface font-medium text-text-primary"
          : "text-text-secondary hover:bg-surface hover:text-text-primary"
      }`}
    >
      {icon}
      {label}
    </Link>
  );
}

export function DisabledNavItem({ icon, label }: { icon: ReactNode; label: string }) {
  return (
    <span
      title="Coming soon"
      className="flex cursor-not-allowed items-center gap-3 rounded-lg px-3 py-2 text-[15px] text-text-muted opacity-50"
    >
      {icon}
      {label}
    </span>
  );
}

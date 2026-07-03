import Link from "next/link";
import type { ReactNode } from "react";

export function EmptyState({
  icon,
  title,
  subtitle,
  ctaHref,
  ctaLabel,
}: {
  icon?: ReactNode;
  title: string;
  subtitle?: string;
  ctaHref?: string;
  ctaLabel?: string;
}) {
  return (
    <div className="flex flex-col items-center gap-3 px-6 py-16 text-center">
      {icon && <div className="text-text-muted">{icon}</div>}
      <p className="text-[15px] font-medium text-text-secondary">{title}</p>
      {subtitle && <p className="text-sm text-text-muted">{subtitle}</p>}
      {ctaHref && ctaLabel && (
        <Link
          href={ctaHref}
          className="mt-1 rounded-full border border-border px-4 py-1.5 text-sm text-text-secondary transition-colors hover:border-accent hover:text-accent"
        >
          {ctaLabel}
        </Link>
      )}
    </div>
  );
}

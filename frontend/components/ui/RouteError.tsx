"use client";

import { useEffect } from "react";

export function RouteError({
  error,
  reset,
  title = "Something went wrong",
}: {
  error: Error & { digest?: string };
  reset: () => void;
  title?: string;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="flex flex-col items-center gap-3 px-6 py-20 text-center">
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className="text-danger">
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="8" x2="12" y2="12" />
        <line x1="12" y1="16" x2="12.01" y2="16" />
      </svg>
      <p className="text-[15px] font-medium text-text-primary">{title}</p>
      <p className="max-w-xs text-sm text-text-muted">
        This didn&apos;t load right. It&apos;s probably transient — try again.
      </p>
      <button
        type="button"
        onClick={reset}
        className="mt-1 rounded-full bg-accent px-5 py-1.5 text-sm font-medium text-white transition-opacity hover:opacity-90"
      >
        Try again
      </button>
    </div>
  );
}

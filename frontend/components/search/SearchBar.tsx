"use client";

import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { useRef, useTransition } from "react";
import { Search, Loader2 } from "lucide-react";

export function SearchBar({ initialQ = "" }: { initialQ?: string }) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [isPending, startTransition] = useTransition();
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const q = e.target.value;
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => {
      const params = new URLSearchParams(searchParams.toString());
      if (q.trim()) {
        params.set("q", q.trim());
      } else {
        params.delete("q");
      }
      // Reset type to all when query changes.
      if (!params.has("type")) params.set("type", "all");
      startTransition(() => {
        router.push(`${pathname}?${params.toString()}`);
      });
    }, 300);
  }

  return (
    <div className="relative">
      <Search
        size={16}
        className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted"
      />
      <input
        type="search"
        defaultValue={initialQ}
        onChange={handleChange}
        placeholder="Search posts and agents…"
        className="w-full rounded-xl border border-border bg-surface py-2.5 pl-9 pr-4 text-[15px] text-text-primary placeholder:text-text-muted focus:border-accent focus:outline-none"
      />
      {isPending && (
        <Loader2
          size={16}
          className="absolute right-3 top-1/2 -translate-y-1/2 animate-spin text-text-muted"
        />
      )}
    </div>
  );
}

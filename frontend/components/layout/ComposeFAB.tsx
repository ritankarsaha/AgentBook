"use client";

import { useState } from "react";
import { PenSquare } from "lucide-react";
import { ComposeModal } from "@/components/feed/ComposeModal";
import type { User } from "@/lib/api";

export function ComposeFAB({ user }: { user: User | null }) {
  const [open, setOpen] = useState(false);

  if (!user) return null;

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        aria-label="New post"
        className="fixed bottom-20 right-4 z-30 flex h-14 w-14 items-center justify-center rounded-full bg-accent text-white shadow-lg transition-all duration-150 hover:scale-110 hover:opacity-90 active:scale-95 lg:hidden"
      >
        <PenSquare size={22} />
      </button>

      {open && <ComposeModal user={user} onClose={() => setOpen(false)} />}
    </>
  );
}

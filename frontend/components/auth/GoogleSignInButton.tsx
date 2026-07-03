"use client";

import { useState } from "react";
import { createClient } from "@/lib/supabase/client";
import { GoogleIcon } from "./GoogleIcon";

const VARIANT_CLASS: Record<string, string> = {
  // Light pill on dark background — matches the splash page treatment.
  "pill-light":
    "w-full flex items-center justify-center gap-3 rounded-full bg-white px-5 py-3 text-sm font-semibold text-black transition-all duration-150 hover:scale-[1.02] hover:bg-white/90 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60",
  // Bordered dark button — matches the existing /login card.
  "outline-dark":
    "mt-6 flex w-full items-center justify-center gap-3 rounded-lg border border-border bg-bg px-4 py-2.5 text-sm font-medium text-text-primary transition-colors hover:bg-border disabled:cursor-not-allowed disabled:opacity-60",
};

export function GoogleSignInButton({
  variant = "outline-dark",
}: {
  variant?: "pill-light" | "outline-dark";
}) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function signInWithGoogle() {
    setLoading(true);
    setError(null);

    const supabase = createClient();
    const { error } = await supabase.auth.signInWithOAuth({
      provider: "google",
      options: {
        redirectTo: `${window.location.origin}/callback`,
      },
    });

    if (error) {
      setError(error.message);
      setLoading(false);
    }
  }

  return (
    <div>
      <button onClick={signInWithGoogle} disabled={loading} className={VARIANT_CLASS[variant]}>
        <GoogleIcon />
        {loading ? "Redirecting…" : "Continue with Google"}
      </button>

      {error && (
        <p className="mt-3 text-sm text-danger" role="alert">
          {error}
        </p>
      )}
    </div>
  );
}

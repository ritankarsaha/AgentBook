"use client";

import { useState } from "react";
import { followHandle, getServerToken, unfollowHandle } from "@/lib/api";
import { useToast } from "@/components/ui/Toast";

export function FollowButton({ handle }: { handle: string }) {
  const [following, setFollowing] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const toast = useToast();

  async function toggle() {
    if (loading) return;
    const token = await getServerToken();
    if (!token) {
      setError("Sign in to follow.");
      toast.error("Sign in to follow.");
      return;
    }
    setLoading(true);
    setError(null);
    try {
      if (following) {
        await unfollowHandle(handle, token);
        setFollowing(false);
        toast.success(`Unfollowed @${handle}`);
      } else {
        await followHandle(handle, token);
        setFollowing(true);
        toast.success(`Following @${handle}`);
      }
    } catch {
      setError("Couldn't update follow — try again.");
      toast.error("Couldn't update follow — try again.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex flex-col items-end gap-1">
      <button
        type="button"
        onClick={toggle}
        disabled={loading}
        className={`rounded-full border px-5 py-1.5 text-sm font-medium transition-all duration-150 hover:scale-105 active:scale-95 disabled:opacity-50 disabled:hover:scale-100 ${
          following
            ? "border-border bg-surface text-text-primary hover:border-danger hover:text-danger"
            : "border-accent bg-accent text-white hover:opacity-90"
        }`}
      >
        {loading ? "…" : following ? "Following" : "Follow"}
      </button>
      {error && <p className="text-xs text-danger">{error}</p>}
    </div>
  );
}

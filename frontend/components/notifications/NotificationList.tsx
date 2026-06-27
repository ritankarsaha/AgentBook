"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import Link from "next/link";
import Image from "next/image";
import { createClient } from "@/lib/supabase/client";
import { getNotifications, getServerToken } from "@/lib/api";
import type { Notification } from "@/lib/api";

const TYPE_LABELS: Record<string, string> = {
  like: "liked your post",
  reply: "replied to your post",
  repost: "reposted your post",
  quote: "quoted your post",
  follow: "followed you",
  mention: "mentioned you",
};

function timeAgo(iso: string): string {
  const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (diff < 60) return `${Math.max(diff, 0)}s`;
  const min = Math.floor(diff / 60);
  if (min < 60) return `${min}m`;
  const hr = Math.floor(min / 60);
  if (hr < 24) return `${hr}h`;
  return `${Math.floor(hr / 24)}d`;
}

function NotifItem({ notif }: { notif: Notification }) {
  const label = TYPE_LABELS[notif.type] ?? notif.type;
  return (
    <div
      className={`flex items-start gap-3 px-4 py-3 transition-colors ${
        !notif.read ? "bg-accent/5" : ""
      }`}
    >
      {/* Avatar */}
      <div className="mt-0.5 h-9 w-9 shrink-0 overflow-hidden rounded-full bg-border">
        {notif.actor_avatar_url ? (
          <Image
            src={notif.actor_avatar_url}
            alt={notif.actor_handle ?? ""}
            width={36}
            height={36}
            className="h-9 w-9 object-cover"
          />
        ) : (
          <div className="flex h-9 w-9 items-center justify-center text-sm text-text-muted">
            {(notif.actor_display_name ?? "?").charAt(0).toUpperCase()}
          </div>
        )}
      </div>

      {/* Body */}
      <div className="min-w-0 flex-1">
        <p className="text-[15px] text-text-primary">
          {notif.actor_handle ? (
            <Link
              href={`/${notif.actor_handle}`}
              className="font-medium hover:underline"
            >
              {notif.actor_display_name ?? notif.actor_handle}
            </Link>
          ) : (
            <span className="font-medium">Someone</span>
          )}{" "}
          <span className="text-text-secondary">{label}</span>
        </p>
        {notif.post_id && (
          <Link
            href={`/posts/${notif.post_id}`}
            className="mt-0.5 block text-sm text-text-muted hover:text-text-secondary"
          >
            View post →
          </Link>
        )}
      </div>

      <span
        className="shrink-0 text-xs text-text-muted"
        suppressHydrationWarning
      >
        {timeAgo(notif.created_at)}
      </span>
    </div>
  );
}

interface NotificationListProps {
  initialNotifs: Notification[];
  initialCursor: string;
  userId: string;
}

type RawNotifRow = { id: string; recipient_user_id?: string; recipient_agent_id?: string };

export function NotificationList({
  initialNotifs,
  initialCursor,
  userId,
}: NotificationListProps) {
  const [notifs, setNotifs] = useState<Notification[]>(initialNotifs);
  const [cursor, setCursor] = useState(initialCursor);
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(!initialCursor);
  const tokenRef = useRef<string | null>(null);
  const sentinel = useRef<HTMLDivElement>(null);

  // Resolve token once.
  useEffect(() => {
    getServerToken()
      .then((t) => {
        tokenRef.current = t;
        // Mark initial notifs as read.
        if (tokenRef.current && initialNotifs.length > 0) {
          getNotifications(tokenRef.current, undefined, 1).catch(() => {});
        }
      });
  }, [initialNotifs.length]);

  const loadMore = useCallback(async () => {
    if (loading || done || !cursor || !tokenRef.current) return;
    setLoading(true);
    try {
      const res = await getNotifications(tokenRef.current, cursor, 30);
      setNotifs((prev) => [...prev, ...(res.data ?? [])]);
      setCursor(res.cursor ?? "");
      if (!res.cursor) setDone(true);
    } catch {
      setDone(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, done, loading]);

  useEffect(() => {
    const el = sentinel.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      (entries) => { if (entries[0].isIntersecting) loadMore(); },
      { rootMargin: "200px" }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [loadMore]);

  // Realtime subscription: new notifications for this user.
  useEffect(() => {
    const supabase = createClient();
    const channel = supabase
      .channel(`notifs-${userId}`)
      .on(
        "postgres_changes",
        {
          event: "INSERT",
          schema: "public",
          table: "notifications",
          filter: `recipient_user_id=eq.${userId}`,
        },
        async (payload: { new: RawNotifRow }) => {
          const row = payload.new;
          if (!tokenRef.current) return;
          try {
            const res = await getNotifications(tokenRef.current, undefined, 1);
            const fresh = res.data?.[0];
            if (fresh && fresh.id === row.id) {
              setNotifs((prev) => [fresh, ...prev.filter((n) => n.id !== fresh.id)]);
            }
          } catch {
            // ignore
          }
        }
      )
      .subscribe();

    return () => { supabase.removeChannel(channel); };
  }, [userId]);

  if (notifs.length === 0 && !loading) {
    return (
      <p className="py-16 text-center text-sm text-text-muted">
        No notifications yet.
      </p>
    );
  }

  return (
    <div>
      <div className="divide-y divide-border">
        {notifs.map((n) => <NotifItem key={n.id} notif={n} />)}
      </div>
      <div ref={sentinel} className="h-4" />
      {loading && <p className="py-4 text-center text-sm text-text-muted">Loading…</p>}
      {done && notifs.length > 0 && (
        <p className="py-6 text-center text-xs text-text-muted">All notifications loaded.</p>
      )}
    </div>
  );
}

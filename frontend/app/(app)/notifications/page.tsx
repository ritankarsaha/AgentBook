import { redirect } from "next/navigation";
import type { Metadata } from "next";
import { createClient } from "@/lib/supabase/server";
import { apiGet } from "@/lib/api";
import { NotificationList } from "@/components/notifications/NotificationList";
import type { Notification } from "@/lib/api";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Notifications",
  robots: { index: false, follow: false },
};

export default async function NotificationsPage() {
  const supabase = await createClient();
  const {
    data: { session },
  } = await supabase.auth.getSession();

  if (!session) {
    redirect("/login");
  }

  let initialNotifs: Notification[] = [];
  let initialCursor = "";
  try {
    const res = await apiGet<Notification[]>(
      "/api/v1/notifications",
      { limit: 30, mark_read: "false" },
      session.access_token
    );
    initialNotifs = res.data ?? [];
    initialCursor = res.cursor ?? "";
  } catch {
    // Show empty state on error.
  }

  return (
    <main className="mx-auto max-w-2xl">
      <h1 className="sticky top-0 z-10 border-b border-border bg-bg/90 px-4 py-3 text-lg font-semibold text-text-primary backdrop-blur">
        Notifications
      </h1>
      <NotificationList
        initialNotifs={initialNotifs}
        initialCursor={initialCursor}
        userId={session.user.id}
      />
    </main>
  );
}

import type { Metadata } from "next";
import { AgentProfileHeader } from "@/components/agent/AgentProfileHeader";
import { UserProfileHeader } from "@/components/user/UserProfileHeader";
import { ProfileFeed } from "@/components/feed/ProfileFeed";
import { createClient } from "@/lib/supabase/server";
import { getAgentPosts, getUserPosts } from "@/lib/api";
import { resolveProfile } from "@/lib/profile";
import type { ProfileTab, Post } from "@/lib/api";

export const dynamic = "force-dynamic";

interface Props {
  params: Promise<{ handle: string }>;
  searchParams: Promise<{ tab?: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { handle } = await params;
  const profile = await resolveProfile(handle);

  if (profile.type === "agent") {
    const a = profile.agent;
    const title = `${a.display_name} (@${a.handle})`;
    const description =
      a.description ||
      `${a.display_name} is an AI agent on AgentBook — ${a.follower_count} followers, ${a.post_count} posts.`;
    const ogImage = `/api/og?${new URLSearchParams({
      title: a.display_name,
      subtitle: `@${a.handle} · ${a.follower_count} followers · ${a.post_count} posts`,
      badge: a.is_verified ? "🤖 verified agent" : "🤖 agent",
    })}`;
    return {
      title,
      description,
      openGraph: { title, description, type: "profile", images: [{ url: ogImage, width: 1200, height: 630 }] },
      twitter: { card: "summary_large_image", title, description, images: [ogImage] },
    };
  }

  const u = profile.user;
  const title = `${u.display_name} (@${u.handle})`;
  const description =
    u.bio || `${u.display_name} on AgentBook — ${u.follower_count} followers, ${u.post_count} posts.`;
  const ogImage = `/api/og?${new URLSearchParams({
    title: u.display_name,
    subtitle: `@${u.handle} on AgentBook`,
  })}`;
  return {
    title,
    description,
    openGraph: { title, description, type: "profile", images: [{ url: ogImage, width: 1200, height: 630 }] },
    twitter: { card: "summary_large_image", title, description, images: [ogImage] },
  };
}

export default async function ProfilePage({ params, searchParams }: Props) {
  const { handle } = await params;
  const { tab: tabParam } = await searchParams;

  // Agents support Posts | Replies | Traces. Humans only Posts | Replies.
  const profile = await resolveProfile(handle);
  const profileType = profile.type;
  const agentProfile = profile.type === "agent" ? profile.agent : null;
  const userProfile = profile.type === "user" ? profile.user : null;

  // Valid tabs differ by profile type.
  const validAgentTabs: ProfileTab[] = ["posts", "replies", "traces"];
  const validUserTabs: ProfileTab[] = ["posts", "replies"];
  const validTabs = profileType === "agent" ? validAgentTabs : validUserTabs;
  const tab: ProfileTab = validTabs.includes(tabParam as ProfileTab)
    ? (tabParam as ProfileTab)
    : "posts";

  // Resolve current user for isOwner.
  const supabase = await createClient();
  const {
    data: { user: authUser },
  } = await supabase.auth.getUser();

  let isOwner = false;
  if (agentProfile && authUser) {
    isOwner = agentProfile.owner_user_id === authUser.id;
  } else if (userProfile && authUser) {
    isOwner = userProfile.id === authUser.id;
  }

  // Fetch first page of posts.
  let initialPosts: Post[] = [];
  let initialCursor = "";
  try {
    const fn = profileType === "agent" ? getAgentPosts : getUserPosts;
    const res = await fn(handle, tab);
    initialPosts = res.data ?? [];
    initialCursor = res.cursor ?? "";
  } catch {
    // Non-fatal — show empty state.
  }

  return (
    <div className="min-h-screen">
      {agentProfile ? (
        <AgentProfileHeader agent={agentProfile} isOwner={isOwner} />
      ) : userProfile ? (
        <UserProfileHeader user={userProfile} isOwner={isOwner} />
      ) : null}

      <ProfileFeed
        key={tab}
        handle={handle}
        profileType={profileType}
        initialPosts={initialPosts}
        initialCursor={initialCursor}
        tab={tab}
        availableTabs={validTabs}
      />
    </div>
  );
}

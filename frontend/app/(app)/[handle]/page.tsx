import { notFound } from "next/navigation";
import { AgentProfileHeader } from "@/components/agent/AgentProfileHeader";
import { UserProfileHeader } from "@/components/user/UserProfileHeader";
import { ProfileFeed } from "@/components/feed/ProfileFeed";
import { createClient } from "@/lib/supabase/server";
import { getAgentProfile, getUserProfile, getAgentPosts, getUserPosts, ApiError } from "@/lib/api";
import type { ProfileTab, Post } from "@/lib/api";

export const dynamic = "force-dynamic";

interface Props {
  params: Promise<{ handle: string }>;
  searchParams: Promise<{ tab?: string }>;
}

export default async function ProfilePage({ params, searchParams }: Props) {
  const { handle } = await params;
  const { tab: tabParam } = await searchParams;

  // Agents support Posts | Replies | Traces. Humans only Posts | Replies.
  let profileType: "agent" | "user" = "agent";
  let agentProfile = null;
  let userProfile = null;

  try {
    const res = await getAgentProfile(handle);
    agentProfile = res.data;
  } catch (e) {
    if (e instanceof ApiError && e.status === 404) {
      profileType = "user";
    } else {
      profileType = "user";
    }
  }

  if (profileType === "user") {
    try {
      const res = await getUserProfile(handle);
      userProfile = res.data;
    } catch {
      notFound();
    }
  }

  if (!agentProfile && !userProfile) notFound();

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

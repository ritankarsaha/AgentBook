-- Row Level Security
-- Agent API key auth is validated in Go (service role key bypasses RLS for
-- agent-authenticated writes). These policies protect the anon/browser path
-- and any direct PostgREST access.

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE agents ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_capabilities ENABLE ROW LEVEL SECURITY;
ALTER TABLE posts ENABLE ROW LEVEL SECURITY;
ALTER TABLE reactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE follows ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_verification_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE platform_stats_daily ENABLE ROW LEVEL SECURITY;

-- users: public profile fields readable by all; only the user themselves can modify
CREATE POLICY users_select_all ON users FOR SELECT USING (true);
CREATE POLICY users_update_self ON users FOR UPDATE USING (auth.uid() = id);

-- agents: directory is public; only the owning user can modify/delete
CREATE POLICY agents_select_all ON agents FOR SELECT USING (true);
CREATE POLICY agents_update_owner ON agents FOR UPDATE USING (auth.uid() = owner_user_id);
CREATE POLICY agents_delete_owner ON agents FOR DELETE USING (auth.uid() = owner_user_id);

CREATE POLICY agent_capabilities_select_all ON agent_capabilities FOR SELECT USING (true);

-- posts: public feed is public; only the author can delete their own post
CREATE POLICY posts_select_all ON posts FOR SELECT USING (true);
CREATE POLICY posts_delete_author ON posts FOR DELETE USING (
  auth.uid() = author_user_id
  OR author_agent_id IN (SELECT id FROM agents WHERE owner_user_id = auth.uid())
);

-- reactions: public counts; reactor can remove their own reaction
CREATE POLICY reactions_select_all ON reactions FOR SELECT USING (true);
CREATE POLICY reactions_delete_self ON reactions FOR DELETE USING (
  auth.uid() = reactor_user_id
  OR reactor_agent_id IN (SELECT id FROM agents WHERE owner_user_id = auth.uid())
);

-- follows: public graph; follower can remove their own follow
CREATE POLICY follows_select_all ON follows FOR SELECT USING (true);
CREATE POLICY follows_delete_self ON follows FOR DELETE USING (
  auth.uid() = follower_user_id
  OR follower_agent_id IN (SELECT id FROM agents WHERE owner_user_id = auth.uid())
);

-- agent_verification_tokens: never exposed to anon/browser, Go service role only
CREATE POLICY verification_tokens_none ON agent_verification_tokens FOR SELECT USING (false);

-- notifications: only the recipient can read their own
CREATE POLICY notifications_select_own ON notifications FOR SELECT USING (
  auth.uid() = recipient_user_id
  OR recipient_agent_id IN (SELECT id FROM agents WHERE owner_user_id = auth.uid())
);
CREATE POLICY notifications_update_own ON notifications FOR UPDATE USING (
  auth.uid() = recipient_user_id
  OR recipient_agent_id IN (SELECT id FROM agents WHERE owner_user_id = auth.uid())
);

-- platform_stats_daily: public read, no anon writes
CREATE POLICY platform_stats_select_all ON platform_stats_daily FOR SELECT USING (true);

CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX idx_posts_author_agent ON posts(author_agent_id, created_at DESC);
CREATE INDEX idx_posts_poster_type ON posts(poster_type, created_at DESC);
CREATE INDEX idx_posts_reply_to ON posts(reply_to_id);
CREATE INDEX idx_posts_engagement ON posts(engagement_score DESC);
CREATE INDEX idx_posts_fts ON posts USING GIN(to_tsvector('english', content));
CREATE INDEX idx_agents_fts ON agents USING GIN(
  to_tsvector('english', handle || ' ' || display_name || ' ' || COALESCE(description, ''))
);
CREATE INDEX idx_agents_handle ON agents(handle);
CREATE INDEX idx_follows_follower_user ON follows(follower_user_id);
CREATE INDEX idx_follows_follower_agent ON follows(follower_agent_id);
CREATE INDEX idx_follows_followee_user ON follows(followee_user_id);
CREATE INDEX idx_follows_followee_agent ON follows(followee_agent_id);
CREATE INDEX idx_notifications_recipient_user ON notifications(recipient_user_id, created_at DESC) WHERE recipient_user_id IS NOT NULL;
CREATE INDEX idx_notifications_recipient_agent ON notifications(recipient_agent_id, created_at DESC) WHERE recipient_agent_id IS NOT NULL;

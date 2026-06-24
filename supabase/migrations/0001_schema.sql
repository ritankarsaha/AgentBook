-- AgentThreads core schema

-- users.id intentionally equals auth.users.id (set by Go backend from the
-- validated JWT sub claim on first sign-in) so RLS can key off auth.uid()
-- directly without a lookup join.
CREATE TABLE users (
  id           UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  email        TEXT UNIQUE NOT NULL,
  handle       TEXT UNIQUE NOT NULL,
  display_name TEXT NOT NULL,
  avatar_url   TEXT,
  bio          TEXT,
  is_verified  BOOLEAN DEFAULT false,
  created_at   TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE agents (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  handle          TEXT UNIQUE NOT NULL,
  display_name    TEXT NOT NULL,
  description     TEXT,
  model           TEXT NOT NULL,
  framework       TEXT,
  api_key_hash    TEXT NOT NULL,
  is_verified     BOOLEAN DEFAULT false,
  verification_badge TEXT,
  avatar_url      TEXT,
  website_url     TEXT,
  agentreplay_id  TEXT,
  last_active_at  TIMESTAMPTZ,
  post_count      INT DEFAULT 0,
  follower_count  INT DEFAULT 0,
  following_count INT DEFAULT 0,
  created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE agent_capabilities (
  agent_id    UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
  capability  TEXT NOT NULL,
  PRIMARY KEY (agent_id, capability)
);

CREATE TABLE posts (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  author_user_id  UUID REFERENCES users(id) ON DELETE SET NULL,
  author_agent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
  poster_type     TEXT NOT NULL CHECK (poster_type IN ('human', 'agent')),
  content         TEXT NOT NULL CHECK (char_length(content) <= 500),
  reply_to_id     UUID REFERENCES posts(id) ON DELETE SET NULL,
  repost_of_id    UUID REFERENCES posts(id) ON DELETE SET NULL,
  quote_content   TEXT,
  media_urls      TEXT[],
  post_subtype    TEXT DEFAULT 'standard',
  trace_url       TEXT,
  like_count      INT DEFAULT 0,
  reply_count     INT DEFAULT 0,
  repost_count    INT DEFAULT 0,
  engagement_score FLOAT GENERATED ALWAYS AS
    (like_count * 3.0 + repost_count * 5.0 + reply_count * 2.0) STORED,
  created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE reactions (
  post_id          UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  reactor_user_id  UUID REFERENCES users(id) ON DELETE CASCADE,
  reactor_agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
  reactor_type     TEXT NOT NULL CHECK (reactor_type IN ('human', 'agent')),
  type             TEXT NOT NULL DEFAULT 'like',
  created_at       TIMESTAMPTZ DEFAULT now(),
  UNIQUE (post_id, reactor_user_id),
  UNIQUE (post_id, reactor_agent_id)
);

CREATE TABLE follows (
  follower_user_id   UUID REFERENCES users(id) ON DELETE CASCADE,
  follower_agent_id  UUID REFERENCES agents(id) ON DELETE CASCADE,
  follower_type      TEXT NOT NULL CHECK (follower_type IN ('human', 'agent')),
  followee_user_id   UUID REFERENCES users(id) ON DELETE CASCADE,
  followee_agent_id  UUID REFERENCES agents(id) ON DELETE CASCADE,
  followee_type      TEXT NOT NULL CHECK (followee_type IN ('human', 'agent')),
  created_at         TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE agent_verification_tokens (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agent_id    UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
  token       TEXT NOT NULL,
  answer      TEXT NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  used        BOOLEAN DEFAULT false,
  created_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE notifications (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  recipient_user_id  UUID REFERENCES users(id) ON DELETE CASCADE,
  recipient_agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
  type               TEXT NOT NULL,
  actor_user_id      UUID REFERENCES users(id),
  actor_agent_id     UUID REFERENCES agents(id),
  post_id            UUID REFERENCES posts(id),
  read               BOOLEAN DEFAULT false,
  created_at         TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE platform_stats_daily (
  date                DATE PRIMARY KEY DEFAULT CURRENT_DATE,
  registered_agents   INT DEFAULT 0,
  active_agents_7d    INT DEFAULT 0,
  human_users         INT DEFAULT 0,
  posts_total         INT DEFAULT 0,
  posts_today         INT DEFAULT 0,
  api_calls_today     INT DEFAULT 0
);

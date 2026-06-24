-- Reserved handles: checked by Go on /api/v1/agents/register before insert.
CREATE TABLE reserved_handles (
  handle TEXT PRIMARY KEY
);

INSERT INTO reserved_handles (handle) VALUES
  ('openai'), ('anthropic'), ('meta'), ('google'), ('nvidia'),
  ('admin'), ('support'), ('agentthreads'), ('moderator'), ('system');

CREATE MATERIALIZED VIEW trending_posts AS
SELECT
  p.*,
  CASE
    WHEN p.created_at > now() - INTERVAL '1 hour'  THEN p.engagement_score * 3
    WHEN p.created_at > now() - INTERVAL '6 hours' THEN p.engagement_score * 2
    WHEN p.created_at > now() - INTERVAL '24 hours' THEN p.engagement_score * 1.5
    ELSE p.engagement_score
  END AS time_weighted_score
FROM posts p
WHERE p.created_at > now() - INTERVAL '7 days'
ORDER BY time_weighted_score DESC;

-- required for REFRESH MATERIALIZED VIEW CONCURRENTLY
CREATE UNIQUE INDEX idx_trending_posts_id ON trending_posts(id);

CREATE EXTENSION IF NOT EXISTS pg_cron;

SELECT cron.schedule(
  'refresh-trending',
  '* * * * *',
  'REFRESH MATERIALIZED VIEW CONCURRENTLY trending_posts'
);

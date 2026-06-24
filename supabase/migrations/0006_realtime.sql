-- Enable Supabase Realtime (WebSocket change feed) on posts so the frontend
-- can subscribe to INSERTs for a live feed (Phase 1.6).
ALTER PUBLICATION supabase_realtime ADD TABLE posts;

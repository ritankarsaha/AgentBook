-- Add website_url to users table for human profile pages.
ALTER TABLE users ADD COLUMN IF NOT EXISTS website_url TEXT;

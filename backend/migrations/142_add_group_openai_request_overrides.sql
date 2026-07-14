ALTER TABLE groups
ADD COLUMN IF NOT EXISTS openai_request_overrides JSONB NOT NULL DEFAULT '{}'::jsonb;

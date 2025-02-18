CREATE TABLE sync_progress (
    id BIGSERIAL PRIMARY KEY,
    koreader_partial_md5 TEXT NOT NULL,
    percentage REAL NOT NULL,
    progress TEXT,
    koreader_device TEXT NOT NULL,
    koreader_device_id TEXT NOT NULL,
    auth_device_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE sync_progress IS 'This table stores the current progress on book to sync between devices';
COMMENT ON COLUMN sync_progress.koreader_device IS 'Device name from KOReader';
COMMENT ON COLUMN sync_progress.auth_device_name IS 'Device name from KOmpanion';

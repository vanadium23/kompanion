CREATE TABLE sync_highlight (
    id BIGSERIAL PRIMARY KEY,
    koreader_partial_md5 TEXT NOT NULL,
    text TEXT NOT NULL,
    text_hash TEXT NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    page TEXT NOT NULL DEFAULT '',
    chapter TEXT NOT NULL DEFAULT '',
    time BIGINT NOT NULL DEFAULT 0,
    drawer TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT '',
    device_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT sync_highlight_unique UNIQUE (koreader_partial_md5, text_hash)
);

CREATE INDEX sync_highlight_koreader_partial_md5 ON sync_highlight(koreader_partial_md5);

COMMENT ON TABLE sync_highlight IS 'Highlights synced from KOReader devices';
COMMENT ON COLUMN sync_highlight.text_hash IS 'SHA-256 hash of highlight text for deduplication';
COMMENT ON COLUMN sync_highlight.time IS 'Unix timestamp from KOReader';

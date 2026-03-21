CREATE TABLE highlight_annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    koreader_partial_md5 TEXT NOT NULL,

    -- Highlight content
    text TEXT NOT NULL,
    note TEXT,
    page TEXT NOT NULL,
    chapter TEXT,

    -- Highlight metadata
    drawer TEXT,
    color TEXT,

    -- Timestamps
    highlight_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Device info
    koreader_device TEXT NOT NULL,
    koreader_device_id TEXT,
    auth_device_name TEXT NOT NULL,

    -- Deduplication
    highlight_hash TEXT NOT NULL
);

CREATE INDEX idx_highlights_document ON highlight_annotations(koreader_partial_md5);
CREATE INDEX idx_highlights_time ON highlight_annotations(highlight_time DESC);
CREATE UNIQUE INDEX idx_highlights_unique ON highlight_annotations(koreader_partial_md5, highlight_hash);

COMMENT ON TABLE highlight_annotations IS 'Stores book highlights synced from KOReader devices';
COMMENT ON COLUMN highlight_annotations.highlight_hash IS 'MD5 hash of (text:page:timestamp) for deduplication';
COMMENT ON COLUMN highlight_annotations.koreader_partial_md5 IS 'MD5 hash of book document for matching with library';

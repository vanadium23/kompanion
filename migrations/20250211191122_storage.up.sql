-- Storage Tables
CREATE TABLE storage_blob (
    id SERIAL PRIMARY KEY,
    file_path TEXT NOT NULL UNIQUE,
    koreader_partial_md5 TEXT NOT NULL,
    file_data BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX storage_blob_koreader_partial_md5_idx ON storage_blob(koreader_partial_md5);
CREATE INDEX storage_blob_file_path_idx ON storage_blob(file_path);

COMMENT ON TABLE storage_blob IS 'Storage for small sized blobs (tenths of MB)';
COMMENT ON COLUMN storage_blob.koreader_partial_md5 IS 'KOREADER partial MD5 hash, stored only for informational purposes';

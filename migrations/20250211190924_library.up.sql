CREATE TABLE library_book (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storage_file_path TEXT NOT NULL UNIQUE,
    koreader_partial_md5 TEXT NOT NULL UNIQUE,
    storage_cover_path TEXT,
    
    -- Book metadata
    title TEXT NOT NULL,
    author TEXT,
    publisher TEXT,
    year INT,
    isbn TEXT,
    series TEXT,
    language TEXT,
    pages INTEGER,
    summary TEXT,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX library_book_title ON library_book(title);
CREATE INDEX library_book_author ON library_book(author);

COMMENT ON TABLE library_book IS 'Plain table without normalization, because we are not expecting a lot of books';

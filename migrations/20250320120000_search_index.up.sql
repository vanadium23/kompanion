-- Add full-text search index for books
-- Using GIN index with tsvector for efficient text search

-- Create a function to generate search vector with weights
-- Weight A: title, author (highest priority)
-- Weight B: series
-- Weight C: publisher, summary

CREATE OR REPLACE FUNCTION book_search_vector()
RETURNS tsvector AS $$
BEGIN
    RETURN
        setweight(to_tsvector('english', COALESCE(title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(author, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(series, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(publisher, '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(summary, '')), 'C');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Add generated column for search vector
ALTER TABLE library_book ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (book_search_vector()) STORED;

-- Create GIN index for fast full-text search
CREATE INDEX library_book_search_idx ON library_book USING GIN(search_vector);

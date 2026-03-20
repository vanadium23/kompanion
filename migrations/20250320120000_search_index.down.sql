-- Remove full-text search index

DROP INDEX IF EXISTS library_book_search_idx;

ALTER TABLE library_book DROP COLUMN IF EXISTS search_vector;

DROP FUNCTION IF EXISTS book_search_vector();

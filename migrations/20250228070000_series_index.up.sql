-- Add series_index column to library_book table
ALTER TABLE library_book ADD COLUMN series_index DECIMAL(10,2);

COMMENT ON COLUMN library_book.series_index IS 'Position of the book in the series (nullable for unnumbered entries)';

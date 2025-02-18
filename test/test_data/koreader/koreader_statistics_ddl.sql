CREATE TABLE book
            (
                id integer PRIMARY KEY autoincrement,
                title text,
                authors text,
                notes      integer,
                last_open  integer,
                highlights integer,
                pages      integer,
                series text,
                language text,
                md5 text,
                total_read_time  integer,
                total_read_pages integer
            );
CREATE UNIQUE INDEX book_title_authors_md5 ON book(title, authors, md5);
CREATE TABLE page_stat_data
        (
            id_book     integer,
            page        integer NOT NULL DEFAULT 0,
            start_time  integer NOT NULL DEFAULT 0,
            duration    integer NOT NULL DEFAULT 0,
            total_pages integer NOT NULL DEFAULT 0,
            UNIQUE (id_book, page, start_time),
            FOREIGN KEY(id_book) REFERENCES book(id)
        );
CREATE INDEX page_stat_data_start_time ON page_stat_data(start_time);
CREATE TABLE numbers
        (
            number INTEGER PRIMARY KEY
        );
CREATE VIEW page_stat AS
        SELECT id_book, first_page + idx - 1 AS page, start_time, duration / (last_page - first_page + 1) AS duration
        FROM (
            SELECT id_book, page, total_pages, pages, start_time, duration,
                -- First page_number for this page after rescaling single row
                ((page - 1) * pages) / total_pages + 1 AS first_page,
                -- Last page_number for this page after rescaling single row
                max(((page - 1) * pages) / total_pages + 1, (page * pages) / total_pages) AS last_page,
                idx
            FROM page_stat_data
            JOIN book ON book.id = id_book
            -- Duplicate rows for multiple pages as needed (as a result of rescaling)
            JOIN (SELECT number as idx FROM numbers) AS N ON idx <= (last_page - first_page + 1)
        );
/* No STAT tables available */

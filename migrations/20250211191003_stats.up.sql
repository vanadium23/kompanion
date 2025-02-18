CREATE TABLE stats_book (
    koreader_partial_md5 TEXT,
    auth_device_name TEXT NOT NULL,
    title TEXT NOT NULL,
    authors TEXT,
    notes INTEGER DEFAULT 0,
    last_open TIMESTAMPTZ,
    highlights INTEGER DEFAULT 0,
    pages INTEGER,
    series TEXT,
    language TEXT,
    total_read_time INTEGER DEFAULT 0,
    total_read_pages INTEGER DEFAULT 0,
    PRIMARY KEY (koreader_partial_md5, auth_device_name)
);
COMMENT ON TABLE stats_book IS 'Table from KOReader stats plugin';

CREATE TABLE stats_page_stat_data (
    koreader_partial_md5 TEXT,
    auth_device_name TEXT NOT NULL,
    page INTEGER NOT NULL DEFAULT 0,
    start_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    duration INTEGER NOT NULL DEFAULT 0,
    total_pages INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (koreader_partial_md5, auth_device_name) REFERENCES stats_book(koreader_partial_md5, auth_device_name),
    UNIQUE (koreader_partial_md5, page, start_time, auth_device_name)
);

-- Indexes
CREATE INDEX stats_page_stat_data_start_time ON stats_page_stat_data(start_time);
CREATE INDEX stats_page_stat_data_start_time_auth_device_name ON stats_page_stat_data(start_time, auth_device_name);

-- Views
CREATE VIEW stats_page_stat AS
SELECT
    sub.title,
    sub.koreader_partial_md5,
    sub.first_page + gs.idx - 1 AS page,
    sub.start_time,
    sub.duration / (sub.last_page - sub.first_page + 1) AS duration
FROM (
    SELECT
        stats_book.title,
        stats_book.koreader_partial_md5,
        page,
        total_pages,
        pages,
        start_time,
        duration,
        ((page - 1) * pages) / total_pages + 1 AS first_page,
        GREATEST(
            ((page - 1) * pages) / total_pages + 1,
            (page * pages) / total_pages
        ) AS last_page
    FROM stats_page_stat_data
    JOIN stats_book 
        ON stats_book.koreader_partial_md5 = stats_page_stat_data.koreader_partial_md5
        AND stats_book.auth_device_name = stats_page_stat_data.auth_device_name
) sub
CROSS JOIN LATERAL generate_series(
    1,
    GREATEST(1, sub.last_page - sub.first_page + 1)
) AS gs(idx);


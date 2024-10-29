-- Use the FTS5 module to create a full-text searchable table.
CREATE VIRTUAL TABLE IF NOT EXISTS html_index 
    USING fts5(host, path, title, text);

-- Will we want to index different things differently?
CREATE VIRTUAL TABLE IF NOT EXISTS pdf_index
    USING fts5(host, path, page_number, text);

-- For now, keep it simple, for demo purposes.
CREATE VIRTUAL TABLE IF NOT EXISTS site_index
    USING fts5(host, path, text);

CREATE TABLE IF NOT EXISTS metadata
    (
        version TEXT NOT NULL,
        last_updated DATE NOT NULL
    )


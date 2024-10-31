-- name: CreateSiteEntry :one
INSERT INTO site_index (host, path, text) 
    VALUES (?, ?, ?)
    RETURNING *;

-- This could become a AS (SELECT ...) 
-- so the table has the data at creation time?
-- name: SetVersion :exec
INSERT INTO metadata (version, last_updated) 
    VALUES (?, DateTime('now'));

-- name: SearchSiteIndex :many
SELECT * FROM site_index 
    WHERE text MATCH ?
    ORDER BY rank
    LIMIT ?;

-- https://www.sqlitetutorial.net/sqlite-full-text-search/
-- name: SearchSiteIndexSnippets :many
SELECT path, snippet(site_index, 2, '<b>', '</b>', '...', 16)
    FROM site_index 
    WHERE text MATCH ?
    ORDER BY rank
    LIMIT ?;


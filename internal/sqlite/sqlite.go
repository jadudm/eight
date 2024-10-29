package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	search_db "search.eight/internal/sqlite/schemas"
)

//go:embed schema.sql
var ddl string

// FIXME: This may become an interface?
type PackTable struct {
	Context context.Context
	DB      *sql.DB
	Queries *search_db.Queries
}

func CreatePackTable(db_filename string) (*PackTable, error) {

	pt := PackTable{}

	ctx := context.Background()

	// Always add an .sqlite extension to filenames.
	if has_ext := strings.HasSuffix("sqlite", db_filename); !has_ext {
		db_filename = db_filename + ".sqlite"
	}

	// FIXME: Any params to the DB?
	db, err := sql.Open("sqlite3", db_filename)
	db.SetMaxOpenConns(1)
	// https://phiresky.github.io/blog/2020/sqlite-performance-tuning/
	db.Exec("pragma journal_mode = WAL")
	db.Exec("pragma synchronous = normal")
	db.Exec("pragma temp_store = memory")
	db.Exec("pragma mmap_size = 30000000000")
	db.Exec("pragma page_size = 32768")

	if err != nil {
		return nil, err
	}

	// create tables
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return nil, err
	}

	queries := search_db.New(db)

	pt.Context = ctx
	pt.DB = db
	pt.Queries = queries

	return &pt, nil
}

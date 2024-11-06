package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"

	search_db "github.com/jadudm/eight/internal/sqlite/schemas"
)

//go:embed schema.sql
var ddl string

// FIXME: This may become an interface?
type PackTable struct {
	Filename string
	Context  context.Context
	DB       *sql.DB
	Queries  *search_db.Queries
	JSON     map[string]string
}

func CreatePackTable(db_filename string, JSON map[string]string) (*PackTable, error) {

	pt := PackTable{}
	pt.Filename = db_filename
	pt.JSON = JSON

	ctx := context.Background()

	// FIXME: Any params to the DB?
	db, err := sql.Open("sqlite3", db_filename)
	db.SetMaxOpenConns(100)
	// https://phiresky.github.io/blog/2020/sqlite-performance-tuning/
	db.Exec("pragma journal_mode = WAL")
	db.Exec("pragma synchronous = normal")
	db.Exec("pragma temp_store = file")
	db.Exec("pragma temp_store_directory = /home/vcap/app/tmp")
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

func (pt *PackTable) PrepForNetwork() {
	// https://turso.tech/blog/something-you-probably-want-to-know-about-if-youre-using-sqlite-in-golang-72547ad625f1
	db, _ := sql.Open("sqlite3", pt.Filename)
	pt.DB = db
	pt.DB.ExecContext(pt.Context, "PRAGMA wal_checkpoint(TRUNCATE)")
	pt.DB.ExecContext(pt.Context, "VACUUM")
	pt.DB.Close()
}

func SqliteFilename(db_filename string) string {
	// Always add an .sqlite extension to filenames.
	if has_ext := strings.HasSuffix(db_filename, "sqlite"); has_ext {
		zap.L().Debug("not adding .sqlite to filename")
		return db_filename
	} else {
		zap.L().Debug("adding .sqlite to filename")
		db_filename = db_filename + ".sqlite"
		return db_filename
	}
}

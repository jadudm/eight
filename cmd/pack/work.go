package main

import (
	"context"
	"log"
	"runtime"
	"sync"

	"github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/sqlite"
	"github.com/jadudm/eight/internal/sqlite/schemas"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

var databases sync.Map

func (w *PackWorker) Work(ctx context.Context, job *river.Job[common.PackArgs]) error {
	zap.L().Debug("packing")

	obj, err := extractStorage.Get(job.Args.Key)
	if err != nil {
		log.Fatal("PACK cannot get obj from S3", job.Args.Key)
	}
	JSON := obj.GetJson()

	host := JSON["host"]

	if _, ok := databases.Load(host); !ok {
		table, err := sqlite.CreatePackTable(sqlite.SqliteFilename(host), JSON)
		if err != nil {
			log.Println("Could not create pack table for", host)
			log.Fatal(err)
		}
		databases.Store(host, table)
	}

	if _pt, ok := databases.Load(host); ok {
		pt := _pt.(*sqlite.PackTable)
		_, err := pt.Queries.CreateSiteEntry(pt.Context, schemas.CreateSiteEntryParams{
			Host: JSON["host"],
			Path: JSON["path"],
			Text: JSON["content"],
		})
		if err != nil {
			log.Println("Insert into site entry table failed")
			log.Fatal(err)
		}
		ch_finalize <- pt
	}

	// Agressively keep memory clear.
	// GC after packing every message.
	runtime.GC()
	return nil
}

package main

import (
	"context"
	"log"
	"os"

	"github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/sqlite"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

func (w *ServeWorker) Work(ctx context.Context, job *river.Job[common.ServeArgs]) error {

	s, _ := env.Env.GetUserService("serve")

	databases_file_path := s.GetParamString("database_files_path")

	sqlite_filename := sqlite.SqliteFilename(job.Args.Filename)

	zap.L().Debug("received sqlite filename",
		zap.String("filename", sqlite_filename))

	// Writes to the local filesystem.
	destination := databases_file_path + "/" + sqlite_filename
	log.Println(destination, "<-", sqlite_filename, os.Getenv("PWD"))

	// Downloads content to the destination
	err := serveStorage.GetFile(sqlite_filename, destination)

	if err != nil {
		zap.L().Error("could not download sqlite db",
			zap.String("sqlite_filename", sqlite_filename), zap.String("destination", destination))
	}
	return nil
}

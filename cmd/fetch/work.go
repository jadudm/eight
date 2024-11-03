package main

import (
	"context"

	"github.com/riverqueue/river"
)

type FetchArgs struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Path   string `json:"path"`
}

func (FetchArgs) Kind() string {
	return "fetch"
}

type FetchWorker struct {
	river.WorkerDefaults[FetchArgs]
}

func (w *FetchWorker) Work(ctx context.Context, job *river.Job[FetchArgs]) error {
	return nil
}

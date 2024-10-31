package pack

import (
	env "github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/queueing"
	"github.com/jadudm/eight/internal/sqlite"
	"github.com/jadudm/eight/pkg/procs"
	"github.com/riverqueue/river"
)

type PackRequest struct {
	Key string `json:"key"`
}

func NewPackRequest() PackRequest {
	return PackRequest{}
}

func (PackRequest) Kind() string {
	b, _ := env.Env.GetBucket("pack")
	return b.Name
}

func (er PackRequest) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: er.Kind(),
	}
}

type PackRequestJob = river.Job[PackRequest]

type PackRequestWorker struct {
	ExtractStorage procs.Storage
	PackStorage    procs.Storage
	EnqueueClient  *queueing.River
	ChanPackages   chan Package
	ChanFinalize   chan *sqlite.PackTable

	river.WorkerDefaults[PackRequest]
}

type PackWorker = river.Worker[PackRequest]

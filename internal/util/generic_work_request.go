package util

import (
	"github.com/riverqueue/river"
)

type GenericRequest struct {
	Key       string `json:"key"`
	QueueName string
}

func NewGenericRequest() GenericRequest {
	return GenericRequest{}
}

func (g GenericRequest) Kind() string {
	return g.QueueName
}

func (g GenericRequest) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: g.Kind(),
	}
}

package queueing

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

// // To be flexible, and use this across the app...
// // this process takes a channel of maps. It looks at the map, and
// // then routes it to the right River work queue.
// func Enqueue(generic_job chan<- kv.JSON) {

// 	// Set up the River client

// 	for {
// 		// Switch on the job kind

// 	}
// }

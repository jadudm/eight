package extract

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log"

	"github.com/jadudm/eight/internal/api"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/util"
	"github.com/jadudm/eight/pkg/procs"
)

type Extractor struct {
	Raw     map[string]string
	Storage procs.Storage
	Job     *ExtractRequestJob
	Stats   *api.BaseStats
}

type ExtractionFunction func(map[string]string)

func NewExtractor(s procs.Storage, raw map[string]string, job *ExtractRequestJob) *Extractor {
	return &Extractor{
		Raw:     raw,
		Storage: s,
		Job:     job,
		Stats:   api.NewBaseStats("extract"),
	}
}

// func _content_key(host string, old_key string, page_number int) string {
// 	sha1 := sha1.Sum([]byte(fmt.Sprintf("%s/%d", old_key, page_number)))
// 	return fmt.Sprintf("%s/%x.json", host, sha1)
// }

func content_key(host string, old_key string, page_number int) string {
	if page_number == -1 {
		// sha1 := sha1.Sum([]byte(fmt.Sprintf("%s", old_key)))
		//return fmt.Sprintf("%s/%x.json", host, sha1)
		return old_key
	} else {
		sha1 := sha1.Sum([]byte(fmt.Sprintf("%s/%d", old_key, page_number)))
		return fmt.Sprintf("%s/%x.json", host, sha1)
	}
}

func (e *Extractor) Extract(erw *ExtractRequestWorker) {
	cleaned_mime_type := util.CleanedMimeType(e.Raw["content-type"])

	s, _ := env.Env.GetUserService("extract")

	switch cleaned_mime_type {
	case "text/html":
		// This inserts into a named queue, not the queue defined by the struct.
		erw.EnqueueClient.InsertTx(util.GenericRequest{
			Key:       e.Job.Args.Key,
			QueueName: "walk"})
		if s.GetParamBool("extract_html") {
			e.ExtractHtml(erw)
		}
	case "application/pdf":
		if s.GetParamBool("extract_pdf") {
			e.ExtractPdf(erw)
		}
	}
}

func (erw *ExtractRequestWorker) Work(
	ctx context.Context,
	job *ExtractRequestJob,
) error {
	log.Println("EXTRACT", job.Args.Key)

	json_object, err := erw.ObjectStorage.Get(job.Args.Key)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(json_object["path"], json_object["content-type"])
	e := NewExtractor(erw.ObjectStorage, json_object, job)
	e.Extract(erw)
	log.Println("EXTRACT DONE", job.Args.Key)

	return nil
}

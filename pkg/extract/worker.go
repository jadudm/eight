package extract

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log"

	"github.com/jadudm/eight/internal/env"
	kv "github.com/jadudm/eight/internal/kv"
	q "github.com/jadudm/eight/internal/queueing"
)

type ExtractionFunction func(map[string]string)

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

func extract(the_client *q.River, obj kv.Object) {
	mime_type := obj.GetMimeType()
	s, _ := env.Env.GetUserService("extract")

	// log.Println("EXTRACT processing MIME type ", mime_type)
	switch mime_type {
	case "text/html":
		if s.GetParamBool("extract_html") {
			log.Println("EXTRACT HTML")
			extractHtml(the_client, obj)
		}
	case "application/pdf":
		if s.GetParamBool("extract_pdf") {
			log.Println("EXTRACT PDF")
			extractPdf(the_client, obj)
		}
	}
}

func (erw *ExtractRequestWorker) Work(
	ctx context.Context,
	job *ExtractRequestJob,
) error {
	log.Println("EXTRACT", job.Args.Key)
	q_client := q.QueueingClient(
		q.NewRiver(),
		q.NewGenericRequest(),
	)

	fetch_bucket := kv.NewKV("fetch")
	obj, err := fetch_bucket.Get(job.Args.Key)
	if err != nil {
		log.Fatal("EXTRACT could not get get key from fetch bucket ", job.Args.Key)
	}

	extract(q_client, obj)

	log.Println("EXTRACT DONE", job.Args.Key)

	return nil
}

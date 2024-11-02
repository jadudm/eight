package extract

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log"

	"github.com/jadudm/eight/internal/env"
	q "github.com/jadudm/eight/internal/queueing"
	kv "github.com/jadudm/eight/pkg/kv"
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
	cleaned_mime_type := obj.GetMimeType()

	s, _ := env.Env.GetUserService("extract")

	log.Println("EXTRACT processing MIME type ", cleaned_mime_type)
	switch cleaned_mime_type {
	case "text/html":
		if s.GetParamBool("walkabout") {
			the_client.InsertTx(q.GenericRequest{
				Key:       obj.GetKey(),
				QueueName: "walk",
			})
		}
		if s.GetParamBool("extract_html") {
			extractHtml(the_client, obj)
		}
	case "application/pdf":
		if s.GetParamBool("extract_pdf") {
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
		log.Fatal("EXTRACT could not get get key from fetch bucket:", job.Args.Key)
	}

	extract(q_client, obj)

	log.Println("EXTRACT DONE", job.Args.Key)

	return nil
}

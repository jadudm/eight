package extract

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"maps"

	"github.com/johbar/go-poppler"
	"search.eight/internal/api"
	"search.eight/pkg/pack"
	"search.eight/pkg/procs"
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

func (e *Extractor) Extract(erw *ExtractRequestWorker) {
	switch e.Raw["content-type"] {
	case "text/html":
		log.Println("HTML")
	case "application/pdf":
		e.ExtractPdf(erw)
	}
}

func content_key(host string, old_key string, page_number int) string {
	sha1 := sha1.Sum([]byte(fmt.Sprintf("%s/%d", old_key, page_number)))
	return fmt.Sprintf("%s/%x.json", host, sha1)
}

func (e *Extractor) ExtractPdf(erw *ExtractRequestWorker) {
	// func process_pdf_bytes(db string, url string, b []byte) {
	// We need a byte array of the original file.
	raw := e.Raw["raw"]

	decoded, err := base64.URLEncoding.DecodeString(raw)

	if err != nil {
		log.Fatal(err)
	}

	// Delete the raw
	delete(e.Raw, "raw")

	doc, err := poppler.Load(decoded)

	if err != nil {
		fmt.Println("Failed to convert body to Document")
	} else {
		for page_no := 0; page_no < doc.GetNPages(); page_no++ {
			extracted_key := content_key(e.Raw["host"], e.Job.Args.Key, page_no+1)
			page := doc.GetPage(page_no)
			new := make(map[string]string, 0)
			// dst, src
			maps.Copy(new, e.Raw)
			new["content"] = page.Text()
			new["path"] = new["path"] + fmt.Sprintf("?page=%d", page_no+1)
			new["pdf_page_number"] = fmt.Sprintf("%d", page_no+1)
			e.Storage.Store(extracted_key, new)
			page.Close()
			e.Stats.Increment("pages_processed")

			// Queue the next step
			erw.EnqueueClient.Insert(pack.PackRequest{
				Key: extracted_key,
			})
		}
	}
	e.Stats.Increment("documents_processed")
	doc.Close()
}

func (erw *ExtractRequestWorker) Work(
	ctx context.Context,
	job *ExtractRequestJob,
) error {
	log.Println("EXTRACT", job.Args.Key)

	json_object, err := erw.FetchStorage.Get(job.Args.Key)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(json_object["path"], json_object["content-type"])
	e := NewExtractor(erw.ExtractStorage, json_object, job)
	e.Extract(erw)
	log.Println("EXTRACT DONE", job.Args.Key)

	return nil
}

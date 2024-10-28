package extract

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"maps"

	"github.com/johbar/go-poppler"
	"search.eight/pkg/procs"
)

type Extractor struct {
	Raw     map[string]string
	Storage procs.Storage
	Job     *ExtractRequestJob
}

type ExtractionFunction func(map[string]string)

func NewExtractor(s procs.Storage, raw map[string]string, job *ExtractRequestJob) *Extractor {
	return &Extractor{
		Raw:     raw,
		Storage: s,
		Job:     job,
	}
}

func (e *Extractor) Extract() {
	switch e.Raw["content-type"] {
	case "text/html":
		log.Println("HTML")
	case "application/pdf":
		e.ExtractPdf()
	}
}

func (e *Extractor) ExtractPdf() {
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
			// log.Println("processing page", page_no)
			page := doc.GetPage(page_no)
			new := make(map[string]string, 0)
			// dst, src
			maps.Copy(new, e.Raw)
			new["content"] = page.Text()
			new["path"] = new["path"] + fmt.Sprintf("?page=%d", page_no+1)
			e.Storage.Store(pdf_page_path(e.Job, new["path"]), new)
			page.Close()
			ES.Increment("pages_processed")
		}
	}
	ES.Increment("documents_processed")
	doc.Close()
}

func pdf_page_path(job *ExtractRequestJob, path string) string {
	sha1 := sha1.Sum([]byte(job.Args.Host + path))
	return fmt.Sprintf("%s/%x.json", job.Args.Host, sha1)
}

func (erw *ExtractRequestWorker) Work(
	ctx context.Context,
	job *ExtractRequestJob,
) error {
	log.Println("EXTRACT", job.Args.Host, job.Args.Path, job.Args.Key)
	// Always safe to check the stats are ready.
	NewExtractStats()

	// FIXME: Need a way to distinguish between  processing an entire
	// host domain and processing a single page?
	// objects, err := erw.FetchStorage.List(job.Args.Host)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println("Found", len(objects), "objects")

	//for _, o := range objects { // use *o.Key as the path
	json_object, err := erw.FetchStorage.Get(job.Args.Key)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(json_object["path"], json_object["content-type"])
	e := NewExtractor(erw.ExtractStorage, json_object, job)
	e.Extract()
	log.Println("EXTRACT DONE", job.Args.Key)
	// }

	return nil
}

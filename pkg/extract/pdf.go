package extract

import (
	"encoding/base64"
	"fmt"
	"log"
	"maps"

	"github.com/jadudm/eight/internal/util"
	"github.com/jadudm/eight/pkg/pack"
	"github.com/johbar/go-poppler"
)

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
			new["content"] = util.RemoveStopwords(page.Text())

			new["path"] = new["path"] + fmt.Sprintf("#page=%d", page_no+1)
			new["pdf_page_number"] = fmt.Sprintf("%d", page_no+1)
			e.Storage.Store(extracted_key, new)
			page.Close()
			e.Stats.Increment("page_count")

			// Queue the next step
			erw.EnqueueClient.InsertTx(pack.PackRequest{
				Key: extracted_key,
			})
		}
	}
	e.Stats.Increment("document_count")
	doc.Close()
}

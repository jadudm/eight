package extract

import (
	"encoding/base64"
	"fmt"
	"log"
	"maps"

	q "github.com/jadudm/eight/internal/queueing"
	"github.com/jadudm/eight/internal/util"
	kv "github.com/jadudm/eight/pkg/kv"
	"github.com/johbar/go-poppler"
)

func extractPdf(q_client *q.River, obj kv.Object) {
	extract_bucket := kv.NewKV("extract")

	jsonm := obj.GetJson()
	raw := jsonm["raw"]

	decoded, err := base64.URLEncoding.DecodeString(raw)

	if err != nil {
		log.Println("PDF cannot Base64 decode")
		log.Fatal(err)
	}

	// Delete the raw
	delete(jsonm, "raw")

	doc, err := poppler.Load(decoded)

	if err != nil {
		fmt.Println("Failed to convert body to Document")
	} else {
		for page_no := 0; page_no < doc.GetNPages(); page_no++ {
			extracted_key := content_key(jsonm["host"], obj.GetKey(), page_no+1)
			page := doc.GetPage(page_no)
			new := make(map[string]string, 0)
			// dst, src
			maps.Copy(new, jsonm)
			new["content"] = util.RemoveStopwords(page.Text())

			new["path"] = new["path"] + fmt.Sprintf("#page=%d", page_no+1)
			new["pdf_page_number"] = fmt.Sprintf("%d", page_no+1)

			extract_bucket.Store(extracted_key, new)
			page.Close()
			// e.Stats.Increment("page_count")

			q_client.InsertTx(q.GenericRequest{
				Key:       obj.GetKey(),
				QueueName: "pack",
			})
		}
	}

	//e.Stats.Increment("document_count")

	doc.Close()
}

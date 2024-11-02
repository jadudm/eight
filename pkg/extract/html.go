package extract

import (
	"bytes"
	"encoding/base64"
	"log"
	"maps"
	"strings"

	"github.com/PuerkitoBio/goquery"
	kv "github.com/jadudm/eight/internal/kv"
	q "github.com/jadudm/eight/internal/queueing"
	"github.com/jadudm/eight/internal/util"
)

func scrape_sel(sel *goquery.Selection) string {
	content := ""
	txt := sel.Text()
	if len(txt) > 0 {
		repl := strings.ToLower(txt)
		// FIXME: This should be part of a standalone text processing
		// module. Perhaps we don't always want to do this. For example,
		// in the case of multi-lingual scraping.
		repl = util.RemoveStopwords(repl)
		repl += " "
		if len(repl) > 2 {
			content += repl
		}
	}
	return content
}

func extractHtml(q_client *q.River, obj kv.Object) {
	extract_bucket := kv.NewKV("extract")

	jsonm := obj.GetJson()
	raw := jsonm["raw"]

	decoded, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		log.Println("HTML cannot Base64 decode")
		log.Fatal(err)
	}
	// Decoded contains a byte array of the raw HTML
	reader := bytes.NewReader(decoded)
	content := ""

	// Delete the raw
	delete(jsonm, "raw")

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Println("HTML cannot create new document")
		log.Fatal(err)
	}

	doc.Find("p").Each(func(ndx int, sel *goquery.Selection) {
		content += scrape_sel(sel)
	})
	doc.Find("li").Each(func(ndx int, sel *goquery.Selection) {
		content += scrape_sel(sel)
	})
	doc.Find("td").Each(func(ndx int, sel *goquery.Selection) {
		content += scrape_sel(sel)
	})

	// Store everything
	extracted_key := util.CreateS3Key(obj.GetValue("host"), obj.GetValue("path")).Render()
	new := make(map[string]string, 0)
	maps.Copy(new, jsonm)
	new["content"] = util.RemoveStopwords(content)

	extract_bucket.Store(extracted_key, new)

	log.Println("EXTRACT queue job for walking", extracted_key)
	q_client.InsertTx(q.GenericRequest{
		Key:       extracted_key,
		QueueName: "walk",
	})

	log.Println("EXTRACT queue job for packing", extracted_key)
	// Queue the next step
	q_client.InsertTx(q.GenericRequest{
		Key:       extracted_key,
		QueueName: "pack",
	})

}

package main

import (
	"log"
	"maps"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jadudm/eight/internal/common"
	kv "github.com/jadudm/eight/internal/kv"
	"github.com/jadudm/eight/internal/util"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
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

func extractHtml(obj kv.Object) {
	extract_bucket := kv.NewKV("extract")

	jsonm := obj.GetJson()
	rawFilename := jsonm["raw"]
	// This gives us a key to a raw file in S3.
	// The key is a UUID that ends in ".raw"
	// use it for the local file
	fetchStorage.GetFile(rawFilename, rawFilename)
	rawFile, err := os.Open(rawFilename)
	if err != nil {
		zap.L().Error("cannot open tempfile", zap.String("filename", rawFilename))
	}
	defer rawFile.Close()

	//reader := bytes.NewReader(rawFile)
	content := ""

	// Delete the raw
	// (This made sense when it was a huge blob. Now it is a file path.)
	// delete(jsonm, "raw")

	doc, err := goquery.NewDocumentFromReader(rawFile)
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
	extracted_key := util.CreateS3Key(obj.GetValue("host"), obj.GetValue("path"), "json").Render()
	new := make(map[string]string, 0)
	maps.Copy(new, jsonm)
	new["content"] = util.RemoveStopwords(content)

	extract_bucket.Store(extracted_key, new)

	// Enqueue next steps
	zap.L().Info("enqueueing pack", zap.String("key", extracted_key))
	ctx, tx := common.CtxTx(packPool)
	defer tx.Rollback(ctx)
	packClient.InsertTx(ctx, tx, common.PackArgs{
		Key: extracted_key,
	}, &river.InsertOpts{Queue: "pack"})
	if err := tx.Commit(ctx); err != nil {
		zap.L().Panic("cannot commit insert tx",
			zap.String("key", extracted_key))
	}

}

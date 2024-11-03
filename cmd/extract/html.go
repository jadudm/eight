package main

import (
	"bytes"
	"encoding/base64"
	"log"
	"maps"
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

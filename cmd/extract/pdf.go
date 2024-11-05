package main

import (
	"fmt"
	"maps"
	"runtime"

	"github.com/jadudm/eight/internal/common"
	kv "github.com/jadudm/eight/internal/kv"
	"github.com/jadudm/eight/internal/util"
	"github.com/johbar/go-poppler"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

func extractPdf(obj kv.Object) {
	jsonm := obj.GetJson()
	rawFilename := jsonm["raw"]
	// This gives us a key to a raw file in S3.
	// The key is a UUID that ends in ".raw"
	// use it for the local file
	fetchStorage.GetFile(rawFilename, rawFilename)

	doc, err := poppler.Open(rawFilename)

	if err != nil {
		fmt.Println("Failed to convert body to Document")
	} else {
		for page_no := 0; page_no < doc.GetNPages(); page_no++ {

			page_number_anchor := fmt.Sprintf("#page=%d", page_no+1)
			extracted_key := util.CreateS3Key(
				obj.GetValue("host"),
				obj.GetValue("path")+page_number_anchor, "json").Render()

			page := doc.GetPage(page_no)
			new := make(map[string]string, 0)
			// dst, src
			maps.Copy(new, jsonm)
			new["content"] = util.RemoveStopwords(page.Text())

			new["path"] = new["path"] + page_number_anchor
			new["pdf_page_number"] = fmt.Sprintf("%d", page_no+1)

			extractStorage.Store(extracted_key, new)
			page.Close()
			// e.Stats.Increment("page_count")

			// Enqueue next steps
			ctx, tx := common.CtxTx(packPool)
			defer tx.Rollback(ctx)

			packClient.InsertTx(ctx, tx, common.PackArgs{
				Key: extracted_key,
			}, &river.InsertOpts{Queue: "pack"})
			if err := tx.Commit(ctx); err != nil {
				zap.L().Panic("cannot commit insert tx",
					zap.String("key", extracted_key))
			}

			// https://weaviate.io/blog/gomemlimit-a-game-changer-for-high-memory-applications
			// https://stackoverflow.com/questions/38972003/how-to-stop-the-golang-gc-and-trigger-it-manually
			runtime.GC()
		}
	}

	//e.Stats.Increment("document_count")

	doc.Close()
}

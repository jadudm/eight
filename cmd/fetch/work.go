package main

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	common "github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/util"
	"github.com/pingcap/log"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

func host_and_path(job *river.Job[common.FetchArgs]) string {
	var u url.URL
	u.Scheme = job.Args.Scheme
	u.Host = job.Args.Host
	u.Path = job.Args.Path
	return u.String()
}

var last_hit sync.Map

func fetch_page_content(job *river.Job[common.FetchArgs]) (map[string]string, error) {
	url := url.URL{
		Scheme: job.Args.Scheme,
		Host:   job.Args.Host,
		Path:   job.Args.Path,
	}

	zap.L().Debug("checking the hit cache")
	if t, ok := last_hit.Load(job.Args.Host); ok {
		if time.Since(t.(time.Time)) < polite_sleep_milliseconds {
			zap.L().Debug("sleeping",
				zap.String("host", job.Args.Host),
				zap.String("path", job.Args.Path))
			time.Sleep(polite_sleep_milliseconds)
		}
	}
	zap.L().Debug("not in the hit cache")
	last_hit.Store(job.Args.Host, time.Now())

	headResp, err := retryablehttp.Head(url.String())
	if err != nil {
		return nil, err
	}

	contentType := headResp.Header.Get("content-type")
	log.Debug("checking HEAD MIME type", zap.String("content-type", contentType))
	if !util.IsSearchableMimeType(contentType) {
		return nil, fmt.Errorf("non-indexable MIME type: %s", url.String())
	}

	// FIXME
	// This eats RAM.
	get_resp, err := retryablehttp.Get(url.String())
	if err != nil {
		zap.L().Fatal("cannot GET content",
			zap.String("url", url.String()),
		)
	}

	zap.L().Debug("successful GET response")

	// Try copying things into a file...
	// tempFile := uuid.NewString()
	// outFile, err := os.Create(tempFile)
	// defer func() { outFile.Close(); os.Remove(tempFile) }()
	// _, err = io.Copy(outFile, get_resp.Body)

	content, err := io.ReadAll(get_resp.Body)
	get_resp.Body.Close()
	if err != nil {
		zap.L().Fatal("cannot io.ReadAll() response",
			zap.String("url", url.String()),
		)
	}

	response := map[string]string{
		"raw":            base64.URLEncoding.EncodeToString(content),
		"sha1":           fmt.Sprintf("%x", sha1.Sum(content)),
		"content-length": fmt.Sprintf("%d", len(content)),
		"host":           job.Args.Host,
		"path":           job.Args.Path,
	}

	// FIXME
	// There is a texinfo standard library for normalizing content types.
	// Consider using it. I want a simplified string, not utf-8 etc.
	response["content-type"] = util.GetMimeType(response["content-type"])

	zap.L().Debug("content read",
		zap.String("content-length", response["content-length"]),
	)

	// Copy in all of the response headers.
	for k := range get_resp.Header {
		response[strings.ToLower(k)] = get_resp.Header.Get(k)
	}

	return response, nil
}

func (w *FetchWorker) Work(ctx context.Context, job *river.Job[common.FetchArgs]) error {
	// Check the cache.
	// We don't want to do anything if this is in the recently visited cache.
	zap.L().Debug("working", zap.String("url", host_and_path(job)))

	// Will aggressive GC keep us under the RAM limit?
	runtime.GC()

	cache_key := host_and_path(job)
	if _, found := recently_visited_cache.Get(cache_key); found {
		zap.L().Debug("cached", zap.String("key", cache_key))
		return nil
	} else {
		// If it is not cached, we have work to do.
		page_json, err := fetch_page_content(job)
		if err != nil {
			zap.L().Warn("could not fetch page content",
				zap.String("scheme", job.Args.Scheme),
				zap.String("host", job.Args.Host),
				zap.String("path", job.Args.Path),
			)
		}

		// FIXME
		// in the grand scheme, we may at this point want to have a queue for
		// coming back a day or two later. But, in terms of fetching... if you can't
		// get to the content... you're not going to store it. So, this
		// bails without sending it back to the queue (for now)
		if err != nil {
			u := url.URL{
				Scheme: job.Args.Scheme,
				Host:   job.Args.Host,
				Path:   job.Args.Path}
			zap.L().Info("could not fetch content; not requeueing",
				zap.String("url", u.String()))
			return nil
		}

		key := util.CreateS3Key(job.Args.Host, job.Args.Path, "json").Render()
		page_json["key"] = key

		zap.L().Debug("storing", zap.String("key", key))
		err = fetchStorage.Store(key, page_json)
		// We get an error if we can't write to S3
		if err != nil {
			zap.L().Warn("could not store k/v",
				zap.String("key", key),
			)
			return err
		}
		zap.L().Debug("stored", zap.String("key", key))

		// Update the cache
		recently_visited_cache.Set(host_and_path(job), key, 0)

		// Enqueue next steps
		// tx, err := fetchPool.Begin(ctx)
		// if err != nil {
		// 	zap.L().Panic("cannot init tx from pool")
		// }
		// defer tx.Rollback(ctx)

		zap.L().Debug("inserting extract job")
		ctx, tx := common.CtxTx(extractPool)
		defer tx.Rollback(ctx)
		extractClient.InsertTx(ctx, tx, common.ExtractArgs{
			Key: key,
		}, &river.InsertOpts{Queue: "extract"})
		if err := tx.Commit(ctx); err != nil {
			zap.L().Panic("cannot commit insert tx",
				zap.String("key", key))
		}

		zap.L().Debug("Inserting walk job")
		ctx2, tx2 := common.CtxTx(walkPool)
		defer tx2.Rollback(ctx)
		walkClient.InsertTx(ctx2, tx2, common.WalkArgs{
			Key: key,
		}, &river.InsertOpts{Queue: "walk"})
		if err := tx2.Commit(ctx2); err != nil {
			zap.L().Info(err.Error())
			zap.L().Panic("cannot commit insert tx",
				zap.String("key", key))
		}
	}

	zap.L().Info("fetched",
		zap.String("scheme", job.Args.Scheme),
		zap.String("host", job.Args.Host),
		zap.String("path", job.Args.Path))

	return nil
}

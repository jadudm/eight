package main

import (
	"bufio"
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	common "github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/util"
	"github.com/pingcap/log"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// ///////////////////////////////////
// GLOBALS
var last_hit sync.Map
var last_backoff sync.Map

//var worker_id atomic.Uint64

func host_and_path(job *river.Job[common.FetchArgs]) string {
	var u url.URL
	u.Scheme = job.Args.Scheme
	u.Host = job.Args.Host
	u.Path = job.Args.Path
	return u.String()
}

func chunkwiseSHA1(filename string) []byte {

	// Open the file for reading.
	tFile, err := os.Open(filename)
	if err != nil {
		zap.L().Error("could not open temp file for encoding to B64")
	}
	defer tFile.Close()
	// Compute the SHA1 going chunk-by-chunk
	h := sha1.New()
	reader := bufio.NewReader(tFile)
	// FIXME: make this a param in the config.
	chunkSize := 4 * 1024
	bytesRead := 0
	buf := make([]byte, chunkSize)
	for {
		n, err := reader.Read(buf)
		bytesRead += n

		if err != nil {
			if err != io.EOF {
				zap.L().Error("chunk error reading")
			}
			break
		}
		chunk := buf[0:n]
		// https://pkg.go.dev/crypto/sha1#example-New
		io.Writer.Write(h, chunk)
	}

	return h.Sum(nil)
}

func getUrlToFile(u url.URL) (string, int64, []byte) {
	getResponse, err := retryablehttp.Get(u.String())
	if err != nil {
		zap.L().Fatal("cannot GET content",
			zap.String("url", u.String()),
		)
	}
	zap.L().Debug("successful GET response")
	// Create a temporary file to download the HTML to.
	temporaryFilename := uuid.NewString()
	outFile, err := os.Create(temporaryFilename)
	if err != nil {
		zap.L().Panic("cannot create temporary file", zap.String("filename", temporaryFilename))
	}
	defer outFile.Close()

	// Copy the Get Reader to a file Writer
	// Should consume little/no RAM.
	// Destination, Source
	bytesRead, err := io.Copy(outFile, getResponse.Body)
	if err != nil {
		zap.L().Panic("could not copy GET to file",
			zap.String("url", u.String()),
			zap.String("filename", temporaryFilename))
	}
	getResponse.Body.Close()
	// Now, it is in a file.
	// Compute the SHA1
	theSHA := chunkwiseSHA1(temporaryFilename)
	return temporaryFilename, bytesRead, theSHA
}

// func getFilesize(filename string) int64 {
// 	fileInfo, err := os.Stat(filename)
// 	if err != nil {
// 		zap.L().Error("could not get filesize", zap.String("filename", filename))
// 	}
// 	return fileInfo.Size()
// }

func fetch_page_content(job *river.Job[common.FetchArgs]) (map[string]string, error) {
	url := url.URL{
		Scheme: job.Args.Scheme,
		Host:   job.Args.Host,
		Path:   job.Args.Path,
	}

	// This holds us up so that the parallel workers don't spam the host.
	common.BackoffLoop(job.Args.Host, polite_sleep_milliseconds, &last_hit, &last_backoff)

	zap.L().Debug("checking the hit cache")

	headResp, err := retryablehttp.Head(url.String())
	if err != nil {
		return nil, err
	}

	contentType := headResp.Header.Get("content-type")
	log.Debug("checking HEAD MIME type", zap.String("content-type", contentType))
	if !util.IsSearchableMimeType(contentType) {
		return nil, fmt.Errorf("non-indexable MIME type: %s", url.String())
	}

	// Write the raw content to a file.
	tempFilename, bytesRead, theSHA := getUrlToFile(url)
	defer func() { os.Remove(tempFilename) }()

	// Stream that file over to S3
	key := util.CreateS3Key(job.Args.Host, job.Args.Path, "raw").Render()

	fetchStorage.StoreFile(key, tempFilename)

	response := map[string]string{
		"raw":            key,
		"sha1":           fmt.Sprintf("%x", theSHA),
		"content-length": fmt.Sprintf("%d", bytesRead),
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
	// This used to be the GET headers, but... they're hiding.
	// Going to do this for now, because I don't know what we'll need.
	for k := range headResp.Header {
		response[strings.ToLower(k)] = headResp.Header.Get(k)
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
			zap.L().Panic("cannot commit insert tx",
				zap.String("key", key), zap.String("error", err.Error()))
		}
	}

	zap.L().Info("fetched",
		zap.String("scheme", job.Args.Scheme),
		zap.String("host", job.Args.Host),
		zap.String("path", job.Args.Path))

	return nil
}

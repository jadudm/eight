package walk

import (
	"bytes"
	"context"
	"encoding/base64"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jadudm/eight/internal/api"
	"github.com/jadudm/eight/internal/util"
	"github.com/jadudm/eight/pkg/fetch"
)

type Walker struct {
	JSON  map[string]string
	Job   *WalkRequestJob
	Stats *api.BaseStats
	WRW   *WalkRequestWorker
}

type WalkionFunction func(map[string]string)

func (e *Walker) Walk() {
	cleaned_mime_type := util.CleanedMimeType(e.JSON["content-type"])
	switch cleaned_mime_type {
	case "text/html":
		e.WalkHTML()
	case "application/pdf":
		log.Println("PDFs do not walk")
	}
}

func (e *Walker) ExtractLinks() []*url.URL {

	raw := e.JSON["raw"]
	decoded, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		log.Fatal(err)
	}
	reader := bytes.NewReader(decoded)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Fatal(err)
	}

	// Return a unique set
	link_set := make(map[string]bool)

	doc.Find("a[href]").Each(func(ndx int, sel *goquery.Selection) {
		link, exists := sel.Attr("href")

		if exists {
			link_to_crawl, err := e.is_crawlable(link)
			if err != nil {
				log.Println(err)
			} else {
				// log.Println("THE CACHE")
				// for ndx, value := range cache.Values() {
				// 	log.Println(ndx, value)
				// }
				if _, ok := cache.Get(link_to_crawl); ok {
					log.Println("CACHE HIT", link)
				} else {
					if strings.HasPrefix(link_to_crawl, "https") {
						//log.Println("YES", e.JSON["host"], link_to_crawl)
						// Don't hit these again
						cache.Set(link_to_crawl, 0, 0)
						link_set[link_to_crawl] = true
					}
				}
			}
		}
	})

	// Remove all trailing slashes.
	links := make([]*url.URL, 0)
	for link := range link_set {
		link = trimSuffix(link, "/")
		u, err := url.Parse(link)
		if err != nil {
			log.Println("WALK ExtractLinks did a bad with", link)
		}
		links = append(links, u)
	}

	//log.Println("EXTRACTED", links)
	return links
}

func (e *Walker) WalkHTML() {
	// func process_pdf_bytes(db string, url string, b []byte) {
	// We need a byte array of the original file.
	links := e.ExtractLinks()
	log.Println("WALK looking at links", links)
	for _, link := range links {
		// Queue the next step
		log.Println("FETCH ENQ", e.JSON["host"], link)
		e.WRW.EnqueueFetch.InsertTx(fetch.FetchRequest{
			// This forces us not to accidentally change
			// the host, at the risk of hitting the wrong domain
			// with a poorly constructed URL
			Host:   e.JSON["host"],
			Scheme: link.Scheme,
			Path:   link.Path,
		})

	}
}

func to_link(JSON map[string]string) string {
	// FIXME: stop assuming the scheme...
	u, _ := url.Parse(
		"https://" +
			JSON["host"] +
			"/" +
			JSON["path"])
	return u.String()
}

func (wrw *WalkRequestWorker) Work(
	ctx context.Context,
	job *WalkRequestJob,
) error {
	log.Println("WALK", job.Args.Key)

	JSON, err := wrw.FetchStorage.Get(job.Args.Key)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(JSON["path"], JSON["content-type"])

	// If we're here, we already fetched the content.
	// So, add ourselves to the cache. Don't re-crawl ourselves
	// FIXME: figure out if the scheme ends up in the JSON
	cache.Set(to_link(JSON), 0, 0)

	e := &Walker{
		JSON:  JSON,
		Job:   job,
		Stats: api.NewBaseStats("walk"),
		WRW:   wrw,
	}

	e.Walk()
	log.Println("WALK DONE", job.Args.Key)

	return nil
}

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	common "github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/kv"
	"github.com/jadudm/eight/internal/util"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// //////////////////////////////////////
// go_for_a_walk
func go_for_a_walk(JSON kv.JSON) {
	cleaned_mime_type := util.CleanMimeType(JSON["content-type"])
	switch cleaned_mime_type {
	case "text/html":
		walk_html(JSON)
	case "application/pdf":
		log.Println("PDFs do not walk")
	}
}

// //////////////////////////////////////
// extract_links
func extract_links(JSON kv.JSON) []*url.URL {

	raw := JSON["raw"]
	decoded, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		log.Println("WALK cannot Base64 decode")
		log.Fatal(err)
	}
	reader := bytes.NewReader(decoded)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Println("WALK cannot convert to document")
		log.Fatal(err)
	}

	// Return a unique set
	link_set := make(map[string]bool)

	doc.Find("a[href]").Each(func(ndx int, sel *goquery.Selection) {
		link, exists := sel.Attr("href")

		if exists {
			link_to_crawl, err := is_crawlable(JSON, link)
			if err != nil {
				log.Println(err)
			} else {
				if _, ok := expirable_cache.Get(link_to_crawl); ok {
					log.Println("CACHE HIT", link)
				} else {
					if strings.HasPrefix(link_to_crawl, "https") {
						//log.Println("YES", e.JSON["host"], link_to_crawl)
						// Don't hit these again
						expirable_cache.Set(link_to_crawl, 0, 0)
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

// //////////////////////////////////////
// walk_html
func walk_html(JSON kv.JSON) {
	// func process_pdf_bytes(db string, url string, b []byte) {
	// We need a byte array of the original file.
	links := extract_links(JSON)
	log.Println("WALK looking at links", links)
	for _, link := range links {
		// Queue the next step
		log.Println("FETCH ENQ", JSON["host"], link)

		ctx, tx := common.CtxTx(dbPool)
		defer tx.Rollback(ctx)
		zap.L().Debug("inserting fetch job")
		fetchClient.InsertTx(context.Background(), tx, common.FetchArgs{
			Host:   JSON["host"],
			Scheme: link.Scheme,
			Path:   link.Path,
		}, &river.InsertOpts{Queue: "fetch"})
		if err := tx.Commit(ctx); err != nil {
			zap.L().Panic("cannot commit insert tx",
				zap.String("host", link.Host), zap.String("path", link.Path))
		}
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

func is_crawlable(JSON kv.JSON, link string) (string, error) {
	host := JSON["host"]
	// FIXME: we should have the scheme in the host?
	scheme := "https"
	path := JSON["path"]
	base := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}

	// Is the URL at least length 1?
	if len(link) < 1 {
		return "", errors.New("crawler: URL is too short to crawl")
	}

	// Is it a relative URL? Then it is OK.
	if string([]rune(link)[0]) == "/" {
		u, _ := url.Parse(link)
		u = base.ResolveReference(u)
		return u.String(), nil
	}

	lu, err := url.Parse(link)
	if err != nil {
		return "", errors.New("crawler: link does not parse")
	}

	// Does it end in .gov?
	// if bytes.HasSuffix([]byte(lu.Host), []byte("gov")) {
	// 	return "", errors.New("crawler: URL does not end in .gov")
	// }

	pieces := strings.Split(base.Host, ".")
	if len(pieces) < 2 {
		return "", errors.New("crawler: link host has too few pieces")
	} else {
		tld := pieces[len(pieces)-2] + "." + pieces[len(pieces)-1]

		// Does the link contain our TLD?
		if !strings.Contains(lu.Host, tld) {
			return "", errors.New("crawler: link does not contain the TLD")
		}
	}

	// FIXME: There seem to be whitespace URLs coming through. I don't know why.
	// This could be revisited, as it is expensive.
	// Do we still have garbage?
	if !bytes.HasSuffix([]byte(lu.String()), []byte("https")) {
		return "", errors.New("crawler: link does not start with https")
	}
	// Is it pure whitespace?
	if len(strings.Replace(lu.String(), " ", "", -1)) < 5 {
		return "", errors.New("crawler: link too short")
	}
	return lu.String(), nil
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
		return s
	} else {
		return s
	}
}

func (w *WalkWorker) Work(ctx context.Context, job *river.Job[common.WalkArgs]) error {
	obj, err := fetchStorage.Get(job.Args.Key)
	JSON := obj.GetJson()

	if err != nil {
		log.Println("WALK cannot grab fetch object", job.Args.Key)
		log.Fatal(err)
	}
	log.Println(JSON["path"], JSON["content-type"])

	// If we're here, we already fetched the content.
	// So, add ourselves to the cache. Don't re-crawl ourselves
	// FIXME: figure out if the scheme ends up in the JSON
	expirable_cache.Set(to_link(JSON), 0, 0)

	go_for_a_walk(JSON)

	log.Println("WALK DONE", job.Args.Key)
	return nil
}

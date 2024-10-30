package extract

import (
	"bytes"
	"encoding/base64"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"search.eight/internal/util"
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

func (e *Extractor) ExtractHtml(erw *ExtractRequestWorker) string {
	raw := e.Raw["raw"]
	decoded, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		log.Fatal(err)
	}
	// Decoded contains a byte array of the raw HTML
	reader := bytes.NewReader(decoded)
	content := ""

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
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

	return content

}

package walk

import (
	"bytes"
	"errors"
	"net/url"
	"strings"
)

func (e *Walker) is_crawlable(link string) (string, error) {
	host := e.JSON["host"]
	// FIXME: we should have the scheme in the host?
	scheme := "https"
	path := e.JSON["path"]
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

package util

import (
	"bytes"
	"strings"
)

func AtoZOnly(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') {
			result.WriteByte(b)
		}
	}
	return result.String()
}

var mime_types = []string{
	"text/html",
	"text/plain",
	"application/pdf",
	"application/x-sqlite3",
}

func CleanMimeType(mime string) string {
	for _, m := range mime_types {
		if strings.Contains(mime, m) {
			return m
		}
	}
	// The unknown mime type
	return "application/octet-stream"
}

func GetMimeType(path string) string {
	m := map[string]string{
		"json":    "application/json",
		"txt":     "text/plain",
		"md":      "text/plain",
		"pdf":     "application/pdf",
		"sqlite":  "application/x-sqlite3",
		"sqlite3": "application/x-sqlite3",
		// https://www.iana.org/assignments/media-types/application/zstd
		"zstd": "application/zstd",
	}
	for k, v := range m {
		if bytes.HasSuffix([]byte(path), []byte(k)) {
			return v
		}
	}
	return m["json"]
}

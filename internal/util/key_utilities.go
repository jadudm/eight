package util

import (
	"crypto/sha1"
	"fmt"
)

type Key struct {
	Host      string
	Path      string
	Extension string
}

func (k *Key) SHA1() string {
	sha := fmt.Sprintf("%x", sha1.Sum([]byte(k.Host+k.Path)))
	return sha
}

func (k *Key) Render() string {
	return fmt.Sprintf("%s/%s.%s", k.Host, k.SHA1(), k.Extension)
}

func CreateS3Key(host string, path string, ext string) *Key {
	return &Key{
		Host:      host,
		Path:      path,
		Extension: ext,
	}
}

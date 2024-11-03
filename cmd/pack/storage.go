package main

import "github.com/jadudm/eight/internal/kv"

var extractStorage kv.S3
var serveStorage kv.S3

func InitializeStorage() {
	extractStorage = kv.NewKV("extract")
	serveStorage = kv.NewKV("serve")
}

package main

import "github.com/jadudm/eight/internal/kv"

var fetchStorage kv.S3

func InitializeStorage() {
	fetchStorage = kv.NewKV("fetch")
}

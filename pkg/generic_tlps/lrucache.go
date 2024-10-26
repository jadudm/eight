package generic_tlps

import (
	"fmt"
	"log"

	"github.com/nalgeon/redka"
)

func LruCache[T any](cache string, in <-chan T, out chan<- T) {
	opts := redka.Options{
		DriverName: "sqlite",
	}
	db, err := redka.Open(fmt.Sprintf("%s.sqlite", cache), &opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// https://github.com/nalgeon/redka/blob/main/docs/commands/hashes.md
}

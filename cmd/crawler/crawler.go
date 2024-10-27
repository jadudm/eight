package main

import (
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"search.eight/pkg/crawl"
)

func BaseMux() *chi.Mux {
	r := chi.NewMux()

	r.Use(middleware.Logger)
	r.Use(middleware.Heartbeat("/heartbeat"))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("HELO"))
	})

	return r
}

func InsertRandomPages(ch chan *crawl.CrawlRequest) {
	for {
		time.Sleep(time.Duration(rand.IntN(10)) * time.Second)
		t := time.Now()

		cr := crawl.NewCrawlRequest()
		cr.Host = uuid.NewString()
		cr.Path = t.String()
		ch <- &cr
	}
}

func main() {
	ch := make(chan *crawl.CrawlRequest)

	r := BaseMux()
	cli, env := ApiCli(r, ch)

	go crawl.Crawl(ch, env)
	cli.Run()
}

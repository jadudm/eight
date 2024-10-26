package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"search.eight/pkg/crawl"
)

var CRAWL_API_VERSION = "1.0.0"

// Options for the CLI.
type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8888"`
}

type CrawlRequestInput struct {
	Body struct {
		Host string `json:"host" maxLength:"500" doc:"Host of resource"`
		Path string `json:"path" maxLength:"1500" doc:"Path to resource"`
	}
}

type RequstReturn func(ctx context.Context, input *CrawlRequestInput) (*struct{}, error)

func CrawlRequestHandler(ch chan *crawl.CrawlRequest) RequstReturn {
	return func(ctx context.Context, input *CrawlRequestInput) (*struct{}, error) {
		ch <- &crawl.CrawlRequest{
			Host: input.Body.Host,
			Path: input.Body.Path,
		}
		return nil, nil
	}
}

func ApiCli(router *chi.Mux, ch chan *crawl.CrawlRequest) humacli.CLI {
	// Create a new router & API
	// router := chi.NewMux()
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		// Will this layer on top of the router I pass in?
		api := humachi.New(router, huma.DefaultConfig("Crawler API", CRAWL_API_VERSION))

		// Register GET /greeting/{name}
		huma.Register(api, huma.Operation{
			OperationID:   "put-crawl-request",
			Method:        http.MethodPut,
			Path:          "/crawl",
			Summary:       "Request a page crawl",
			Description:   "Request a crawl of a path at a given host.",
			Tags:          []string{"Crawl"},
			DefaultStatus: http.StatusAccepted,
		}, CrawlRequestHandler(ch))

		// Tell the CLI how to start your server.
		hooks.OnStart(func() {
			fmt.Printf("Starting server on port %d...\n", options.Port)
			http.ListenAndServe(fmt.Sprintf(":%d", options.Port), router)
		})
	})

	return cli
}

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

		ch <- &crawl.CrawlRequest{
			Host: uuid.NewString(),
			Path: t.String(),
		}
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	ch := make(chan *crawl.CrawlRequest)

	go crawl.Crawl(ch)
	//go InsertRandomPages(ch)

	r := BaseMux()
	cli := ApiCli(r, ch)
	cli.Run()
}

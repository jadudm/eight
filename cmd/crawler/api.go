package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"search.eight/internal/env"
	"search.eight/pkg/crawl"
)

var CRAWL_API_VERSION = "1.0.0"

// Options for the CLI.
type Options struct {
	Port   int    `help:"Port to listen on" short:"p" default:"8888"`
	Config string `help:"VCAP_SERVICES JSON file" short:"c" default:"vcap.json"`
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

func ApiCli(router *chi.Mux, ch chan *crawl.CrawlRequest) (humacli.CLI, *env.Env) {
	// Create a new router & API
	// router := chi.NewMux()
	var e *env.Env

	e = env.NewFromFile("vcap.json")

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

	return cli, e
}

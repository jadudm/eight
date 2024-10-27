package main

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"search.eight/pkg/fetch"
)

var FETCH_API_VERSION = "1.0.0"

// Options for the CLI.
type Options struct {
	Port   int    `help:"Port to listen on" short:"p" default:"8888"`
	Config string `help:"VCAP_SERVICES JSON file" short:"c" default:"vcap.json"`
}

type FetchRequestInput struct {
	Body struct {
		Host string `json:"host" maxLength:"500" doc:"Host of resource"`
		Path string `json:"path" maxLength:"1500" doc:"Path to resource"`
	}
}

type RequstReturn func(ctx context.Context, input *FetchRequestInput) (*struct{}, error)

func FetchRequestHandler(ch chan *fetch.FetchRequest) RequstReturn {
	return func(ctx context.Context, input *FetchRequestInput) (*struct{}, error) {
		cr := fetch.NewFetchRequest()
		cr.Host = input.Body.Host
		cr.Path = input.Body.Path
		ch <- &cr
		return nil, nil
	}
}

func FetchApi(router *chi.Mux, ch chan *fetch.FetchRequest) *chi.Mux {
	// Will this layer on top of the router I pass in?
	api := humachi.New(router, huma.DefaultConfig("Fetch API", FETCH_API_VERSION))

	// Register GET /greeting/{name}
	huma.Register(api, huma.Operation{
		OperationID:   "put-fetch-request",
		Method:        http.MethodPut,
		Path:          "/fetch",
		Summary:       "Request a page fetch",
		Description:   "Request a fetch of a path at a given host.",
		Tags:          []string{"Fetch"},
		DefaultStatus: http.StatusAccepted,
	}, FetchRequestHandler(ch))

	return router
}

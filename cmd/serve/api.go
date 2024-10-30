package main

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"search.eight/pkg/serve"
)

var SERVE_API_VERSION = "1.0.0"

// FIXME This becomes the API search interface
type ServeRequestInput struct {
	Body struct {
		Host string `json:"host" maxLength:"500" doc:"Host of resource"`
		Path string `json:"path" maxLength:"1500" doc:"Path to resource"`
	}
}

type RequestReturn func(ctx context.Context, input *ServeRequestInput) (*struct{}, error)

func ServeRequestHandler(ch chan *serve.ServeRequest) RequestReturn {
	return func(ctx context.Context, input *ServeRequestInput) (*struct{}, error) {
		cr := serve.NewServeRequest()
		// cr.Host = input.Body.Host
		// cr.Path = input.Body.Path
		ch <- &cr
		return nil, nil
	}
}

func ServeApi(router *chi.Mux, ch chan *serve.ServeRequest) *chi.Mux {
	// Will this layer on top of the router I pass in?
	api := humachi.New(router, huma.DefaultConfig("Serve API", SERVE_API_VERSION))

	// Register GET /greeting/{name}
	huma.Register(api, huma.Operation{
		OperationID:   "put-serve-request",
		Method:        http.MethodPut,
		Path:          "/serve",
		Summary:       "Request a page serve",
		Description:   "Request a serve of a path at a given host.",
		Tags:          []string{"Serve"},
		DefaultStatus: http.StatusAccepted,
	}, ServeRequestHandler(ch))

	return router
}

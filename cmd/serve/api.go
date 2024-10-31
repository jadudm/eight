package main

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"search.eight/pkg/serve"
)

var SERVE_API_VERSION = "1.0.0"

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

	huma.Register(api, huma.Operation{
		OperationID:   "post-serve-request",
		Method:        http.MethodPost,
		Path:          "/serve",
		Summary:       "Search a host",
		Description:   "Search a host",
		Tags:          []string{"Serve"},
		DefaultStatus: http.StatusAccepted,
	}, ServeHandler)

	huma.Register(api, huma.Operation{
		OperationID:   "get-info-request",
		Method:        http.MethodGet,
		Path:          "/databases",
		Summary:       "List the databases available",
		Description:   "List the databases available",
		Tags:          []string{"list"},
		DefaultStatus: http.StatusAccepted,
	}, ListDatabaseRequestHandler)

	return router
}

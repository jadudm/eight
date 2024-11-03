package main

import (
	"context"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jadudm/eight/internal/api"
)

var FETCH_API_VERSION = "1.0.0"

type FetchRequestInput struct {
	Body struct {
		Scheme string `json:"scheme" maxLength:"10" doc:"Resource scheme"`
		Host   string `json:"host" maxLength:"500" doc:"Host of resource"`
		Path   string `json:"path" maxLength:"1500" doc:"Path to resource"`
		ApiKey string `json:"api-key"`
	}
}

type RequestReturn func(ctx context.Context, input *FetchRequestInput) (*struct{}, error)

func InitializeAPI() *chi.Mux {
	r := api.BaseMux()
	api.MemInfo(r)
	AddFetchApi(r)
	return r
}

func FetchRequestHandler(ctx context.Context, input *FetchRequestInput) (*struct{}, error) {
	if input.Body.ApiKey == os.Getenv("API_KEY") {
		// The third parameter to .Insert() is &river.InsertOpts{}.
		// We can use that to override the target queue, if we want.
		fetchClient.Insert(context.Background(), FetchArgs{
			Scheme: input.Body.Scheme,
			Host:   input.Body.Host,
			Path:   input.Body.Path,
		}, nil)
	}
	return nil, nil
}

func AddFetchApi(router *chi.Mux) {
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
	}, FetchRequestHandler)
}

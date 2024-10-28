package main

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"search.eight/pkg/extract"
)

var FETCH_API_VERSION = "1.0.0"

type ExtractRequestInput struct {
	Body struct {
		Host string `json:"host" maxLength:"500" doc:"Host of resource"`
	}
}

type RequestReturn func(ctx context.Context, input *ExtractRequestInput) (*struct{}, error)

func ExtractRequestHandler(ch chan *extract.ExtractRequest) RequestReturn {
	return func(ctx context.Context, input *ExtractRequestInput) (*struct{}, error) {
		er := extract.NewExtractRequest()
		er.Host = input.Body.Host
		ch <- &er
		return nil, nil
	}
}

func ExtractApi(router *chi.Mux, ch chan *extract.ExtractRequest) *chi.Mux {
	// Will this layer on top of the router I pass in?
	api := humachi.New(router, huma.DefaultConfig("Extract API", FETCH_API_VERSION))

	// Register GET /greeting/{name}
	huma.Register(api, huma.Operation{
		OperationID:   "put-extract-request",
		Method:        http.MethodPut,
		Path:          "/extract",
		Summary:       "Request content extraction",
		Description:   "Request extraction of text from a fetched page.",
		Tags:          []string{"Extract"},
		DefaultStatus: http.StatusAccepted,
	}, ExtractRequestHandler(ch))

	return router
}

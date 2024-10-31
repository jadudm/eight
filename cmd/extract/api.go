package main

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jadudm/eight/internal/api"
	"github.com/jadudm/eight/pkg/extract"
)

var EXTRACT_API_VERSION = "1.0.0"

type ExtractRequestInput struct {
	Body struct {
		Key string `json:"key" maxLength:"2000" doc:"Key of object in S3"`
	}
}

type RequestReturn func(ctx context.Context, input *ExtractRequestInput) (*struct{}, error)

func ExtractRequestHandler(ch chan *extract.ExtractRequest) RequestReturn {
	return func(ctx context.Context, input *ExtractRequestInput) (*struct{}, error) {
		er := extract.NewExtractRequest()
		er.Key = input.Body.Key
		ch <- &er
		return nil, nil
	}
}

func ExtractApi(router *chi.Mux, ch chan *extract.ExtractRequest) *chi.Mux {
	// Will this layer on top of the router I pass in?
	huma_api := humachi.New(router, huma.DefaultConfig("Extract API", EXTRACT_API_VERSION))

	// Register GET /meminfo
	huma.Register(huma_api, huma.Operation{
		OperationID:   "put-extract-request",
		Method:        http.MethodPut,
		Path:          "/extract",
		Summary:       "Request content extraction",
		Description:   "Request extraction of text from a fetched page.",
		Tags:          []string{"Extract"},
		DefaultStatus: http.StatusAccepted,
	}, ExtractRequestHandler(ch))

	// Register GET /stats
	huma.Register(huma_api, huma.Operation{
		OperationID:   "get-stats-request",
		Method:        http.MethodGet,
		Path:          "/stats",
		Summary:       "Request stats about this app",
		Description:   "Request stats about this app",
		Tags:          []string{"stats"},
		DefaultStatus: http.StatusAccepted,
	}, api.StatsHandler("extract"))

	return router
}

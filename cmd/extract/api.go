package main

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"search.eight/pkg/extract"
)

var EXTRACT_API_VERSION = "1.0.0"

type StatsInput struct{}
type StatsResponse struct {
	Stats map[string]int64
}

func StatsHandler(ctx context.Context, input *StatsInput) (*StatsResponse, error) {
	// Does nothing if the stats are already initialized.
	extract.NewExtractStats()
	s := extract.ES.GetAll()
	return &StatsResponse{Stats: s}, nil
}

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
	api := humachi.New(router, huma.DefaultConfig("Extract API", EXTRACT_API_VERSION))

	// Register GET /meminfo
	huma.Register(api, huma.Operation{
		OperationID:   "put-extract-request",
		Method:        http.MethodPut,
		Path:          "/extract",
		Summary:       "Request content extraction",
		Description:   "Request extraction of text from a fetched page.",
		Tags:          []string{"Extract"},
		DefaultStatus: http.StatusAccepted,
	}, ExtractRequestHandler(ch))

	// Register GET /stats
	huma.Register(api, huma.Operation{
		OperationID:   "get-stats-request",
		Method:        http.MethodGet,
		Path:          "/stats",
		Summary:       "Request stats about this app",
		Description:   "Request stats about this app",
		Tags:          []string{"stats"},
		DefaultStatus: http.StatusAccepted,
	}, StatsHandler)

	return router
}

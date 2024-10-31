package main

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jadudm/eight/internal/api"
	"github.com/jadudm/eight/pkg/pack"
)

var PACK_API_VERSION = "1.0.0"

type PackRequestInput struct {
	Body struct {
		Key string `json:"host" maxLength:"500" doc:"Key of resource"`
	}
}

type RequestReturn func(ctx context.Context, input *PackRequestInput) (*struct{}, error)

func PackRequestHandler(ch chan *pack.PackRequest) RequestReturn {
	return func(ctx context.Context, input *PackRequestInput) (*struct{}, error) {
		er := pack.NewPackRequest()
		er.Key = input.Body.Key
		ch <- &er
		return nil, nil
	}
}

func PackApi(router *chi.Mux, ch chan *pack.PackRequest) *chi.Mux {
	// Will this layer on top of the router I pass in?
	huma_api := humachi.New(router, huma.DefaultConfig("Pack API", PACK_API_VERSION))

	// Register GET /meminfo
	huma.Register(huma_api, huma.Operation{
		OperationID:   "put-pack-request",
		Method:        http.MethodPut,
		Path:          "/pack",
		Summary:       "Request content packion",
		Description:   "Request packion of text from a fetched page.",
		Tags:          []string{"Pack"},
		DefaultStatus: http.StatusAccepted,
	}, PackRequestHandler(ch))

	// Register GET /stats
	huma.Register(huma_api, huma.Operation{
		OperationID:   "get-stats-request",
		Method:        http.MethodGet,
		Path:          "/stats",
		Summary:       "Request stats about this app",
		Description:   "Request stats about this app",
		Tags:          []string{"stats"},
		DefaultStatus: http.StatusAccepted,
	}, api.StatsHandler("pack"))

	return router
}

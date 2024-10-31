package main

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jadudm/eight/internal/api"
	"github.com/jadudm/eight/pkg/walk"
)

var WALK_API_VERSION = "1.0.0"

type WalkRequestInput struct {
	Body struct {
		Key string `json:"key" maxLength:"2000" doc:"Key of object in S3"`
	}
}

type RequestReturn func(ctx context.Context, input *WalkRequestInput) (*struct{}, error)

func WalkRequestHandler(ch chan *walk.WalkRequest) RequestReturn {
	return func(ctx context.Context, input *WalkRequestInput) (*struct{}, error) {
		er := walk.NewWalkRequest()
		er.Key = input.Body.Key
		ch <- &er
		return nil, nil
	}
}

func WalkApi(router *chi.Mux, ch chan *walk.WalkRequest) *chi.Mux {
	// Will this layer on top of the router I pass in?
	huma_api := humachi.New(router, huma.DefaultConfig("Walk API", WALK_API_VERSION))

	// Register GET /meminfo
	huma.Register(huma_api, huma.Operation{
		OperationID:   "put-walk-request",
		Method:        http.MethodPut,
		Path:          "/walk",
		Summary:       "Request content walk",
		Description:   "Request walk of text from a fetched page.",
		Tags:          []string{"Walk"},
		DefaultStatus: http.StatusAccepted,
	}, WalkRequestHandler(ch))

	// Register GET /stats
	huma.Register(huma_api, huma.Operation{
		OperationID:   "get-stats-request",
		Method:        http.MethodGet,
		Path:          "/stats",
		Summary:       "Request stats about this app",
		Description:   "Request stats about this app",
		Tags:          []string{"stats"},
		DefaultStatus: http.StatusAccepted,
	}, api.StatsHandler("walk"))

	return router
}

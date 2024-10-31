package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"search.eight/internal/sqlite/schemas"
	"search.eight/pkg/serve"
)

var SERVE_API_VERSION = "1.0.0"

// FIXME This becomes the API search interface
type ServeRequestInput struct {
	Body struct {
		Host  string `json:"host" maxLength:"500" doc:"Host to search"`
		Terms string `json:"terms" maxLength:"200" doc:"Search terms"`
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

type ServeResponse struct {
	Result  string                               `json:"result"`
	Elapsed time.Duration                        `json:"elapsed"`
	Results []schemas.SearchSiteIndexSnippetsRow `json:"results"`
}

type ServeResponseBody struct {
	Body *ServeResponse
}

func ServeHandler(ctx context.Context, input *ServeRequestInput) (*ServeResponseBody, error) {
	start := time.Now()
	host := input.Body.Host
	terms := input.Body.Terms

	// Search DB
	//ctx := context.Background()
	sqlite_file := host + ".sqlite"
	db, err := sql.Open("sqlite3", sqlite_file)
	if err != nil {
		log.Fatal("SERVCE cannot open SQLite file", sqlite_file)
	}

	queries := schemas.New(db)
	res, err := queries.SearchSiteIndexSnippets(ctx, schemas.SearchSiteIndexSnippetsParams{
		Text:  terms,
		Limit: 20,
	})

	duration := time.Since(start)

	if err != nil {
		return &ServeResponseBody{
			Body: &ServeResponse{
				Result:  "err",
				Elapsed: duration,
				Results: nil,
			}}, err
	} else {
		return &ServeResponseBody{
			Body: &ServeResponse{
				Result:  "ok",
				Elapsed: duration,
				Results: res,
			}}, nil
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

	huma.Register(api, huma.Operation{
		OperationID:   "post-serve-request",
		Method:        http.MethodPost,
		Path:          "/serve",
		Summary:       "Search a host",
		Description:   "Search a host",
		Tags:          []string{"Serve"},
		DefaultStatus: http.StatusAccepted,
	}, ServeHandler)

	return router
}

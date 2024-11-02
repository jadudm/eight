package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	apistats "github.com/jadudm/eight/internal/api"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/sqlite/schemas"
	"github.com/jadudm/eight/pkg/serve"
)

// FIXME This becomes the API search interface
type ServeRequestInput struct {
	Body struct {
		Host  string `json:"host" maxLength:"500" doc:"Host to search"`
		Terms string `json:"terms" maxLength:"200" doc:"Search terms"`
	}
}

type ServeRequestReturn func(ctx context.Context, input *ServeRequestInput) (*struct{}, error)

func ServeRequestHandler(ch chan *serve.ServeRequest) ServeRequestReturn {
	return func(ctx context.Context, input *ServeRequestInput) (*struct{}, error) {
		cr := serve.NewServeRequest()
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

	s, _ := env.Env.GetUserService("serve")
	database_files_path := s.GetParamString("database_files_path")
	results_per_query := s.GetParamInt64("results_per_query")

	sqlite_file := database_files_path + "/" + host + ".sqlite"
	log.Println(sqlite_file)
	if _, err := os.Stat(sqlite_file); errors.Is(err, os.ErrNotExist) {
		duration := time.Since(start)
		return &ServeResponseBody{
			Body: &ServeResponse{
				Result:  "err",
				Elapsed: duration,
				Results: nil,
			}}, err
	}

	db, err := sql.Open("sqlite3", sqlite_file)
	if err != nil {
		log.Fatal("SERVCE cannot open SQLite file", sqlite_file)
	}

	queries := schemas.New(db)
	res, err := queries.SearchSiteIndexSnippets(ctx, schemas.SearchSiteIndexSnippetsParams{
		Text:  terms,
		Limit: results_per_query,
	})

	duration := time.Since(start)

	// Search accounting
	stats := apistats.NewBaseStats("serve")
	stats.Increment("_queries")
	stats.Increment("_" + input.Body.Host)
	stats.Sum("_total_query_time", duration.Nanoseconds())
	stats.Set("_average_query_time", int64(stats.Get("_total_query_time")/stats.Get("_queries")))

	// Count all the search terms? Why not!
	for _, t := range strings.Split(terms, " ") {
		stats.Increment(t)
	}

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

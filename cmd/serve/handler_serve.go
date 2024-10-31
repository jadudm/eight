package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"time"

	"search.eight/internal/env"
	"search.eight/internal/sqlite/schemas"
	"search.eight/pkg/serve"
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
	s, _ := env.Env.GetService("serve")
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

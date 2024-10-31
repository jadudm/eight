package main

import (
	"context"
	"log"
	"os"
	"strings"

	"search.eight/internal/env"
)

// This is essentially a full API back-to-front.
/*
huma.Register(api, huma.Operation{
	OperationID:   "post-info-request",
	Method:        http.MethodPost,
	Path:          "/info",
	Summary:       "Info about the service",
	Description:   "Info about the service",
	Tags:          []string{"Serve"},
	DefaultStatus: http.StatusAccepted,
}, InfoRequestHandler)

type InfoRequestInput struct {
	Body struct {
		Host  string `json:"host" maxLength:"500" doc:"Host to search"`
		Terms string `json:"terms" maxLength:"200" doc:"Search terms"`
	}
}

type InfoRequestReturn func(ctx context.Context, input *InfoRequestInput) (*struct{}, error)

type InfoResponse struct {
	Result  string                               `json:"result"`
	Elapsed time.Duration                        `json:"elapsed"`
	Results []schemas.SearchSiteIndexSnippetsRow `json:"results"`
}

func InfoRequestHandler(ctx context.Context, input *InfoRequestInput) (*struct{}, error) {

	return nil, nil
}
*/

type LDBRequestInput struct{}

type LDBRequestReturn func(ctx context.Context, input *LDBRequestInput) (*struct{}, error)

type LDBResponse struct {
	Databases []string `json:"databases"`
}

type LDBResponseBody struct {
	Body *LDBResponse
}

func ListDatabaseRequestHandler(ctx context.Context, input *LDBRequestInput) (*LDBResponseBody, error) {
	s, _ := env.Env.GetService("serve")
	database_files_path := s.GetParamString("database_files_path")

	files, err := os.ReadDir(database_files_path)
	if err != nil {
		log.Fatal(err)
	}

	dbs := make([]string, 0)
	suffix := ".sqlite"
	for _, file := range files {
		if strings.HasSuffix(file.Name(), suffix) {
			dbs = append(dbs, strings.TrimSuffix(file.Name(), suffix))
		}
	}

	log.Println("INFO", dbs)
	return &LDBResponseBody{
		Body: &LDBResponse{
			Databases: dbs,
		}}, nil
}

package common

import "net/http"

type ExtractArgs struct {
	Key string `json:"key"`
}

func (ExtractArgs) Kind() string {
	return "extract"
}

type FetchArgs struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Path   string `json:"path"`
}

func (FetchArgs) Kind() string {
	return "fetch"
}

type PackArgs struct {
	Key string `json:"key"`
}

func (PackArgs) Kind() string {
	return "pack"
}

type WalkArgs struct {
	Key string `json:"key"`
}

func (WalkArgs) Kind() string {
	return "extract"
}

type HttpResponse func(w http.ResponseWriter, r *http.Request)

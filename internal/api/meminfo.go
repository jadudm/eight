package api

import (
	"context"
	"net/http"
	"runtime"
	"runtime/debug"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

var MEMINFO_API_VERSION = "1.0.0"

type MemInfoInput struct{}
type MemInfoResponse struct {
	Mem runtime.MemStats
	GC  debug.GCStats
}

func MemInfoHandler(ctx context.Context, input *MemInfoInput) (*MemInfoResponse, error) {
	mem := MemInfoResponse{}
	runtime.ReadMemStats(&mem.Mem)
	debug.ReadGCStats(&mem.GC)
	return &mem, nil
}

func MemInfo(router *chi.Mux) {
	// Will this layer on top of the router I pass in?
	api := humachi.New(router, huma.DefaultConfig("Fetch API", MEMINFO_API_VERSION))

	// Register GET /greeting/{name}
	huma.Register(api, huma.Operation{
		OperationID:   "get-memino-request",
		Method:        http.MethodGet,
		Path:          "/meminfo",
		Summary:       "Request memory info about this app",
		Description:   "Request memory info about this app",
		Tags:          []string{"meminfo"},
		DefaultStatus: http.StatusAccepted,
	}, MemInfoHandler)

}

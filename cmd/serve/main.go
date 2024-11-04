package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/kv"
	"go.uber.org/zap"
)

var serveStorage kv.S3

func ServeHost(c *gin.Context) {
	s, _ := env.Env.GetUserService("serve")
	external_scheme := s.GetParamString("external_scheme")
	external_host := s.GetParamString("external_host")
	external_port := s.GetParamInt64("external_port")
	static_files_path := s.GetParamString("static_files_path")

	zap.L().Debug("serving up a host", zap.String("external_host", external_host))

	host := c.Param("host")
	data, err := os.ReadFile(static_files_path + "/index.html")
	if err != nil {
		log.Println("SERVE could not read index.html")
		log.Fatal(err)
	}
	data = bytes.ReplaceAll(data, []byte("{SCHEME}"), []byte(external_scheme))
	data = bytes.ReplaceAll(data, []byte("{HOST}"), []byte(external_host))
	data = bytes.ReplaceAll(data, []byte("{SEARCH_HOST}"), []byte(host))

	data = bytes.ReplaceAll(data, []byte("{PORT}"), []byte(fmt.Sprintf("%d", external_port)))

	c.Data(http.StatusOK, "text/html", data)
}

func MultiStatsHandler(c *gin.Context) {
	dbs := listHostedDOmains()
	res := make(map[string]any)
	base_stats := common.NewBaseStats("serve")
	res["stats"] = base_stats.GetAll()
	for _, db := range dbs {
		st := common.NewBaseStats(db)
		res[db] = st.GetAll()
	}
	res["hosted_domains"] = dbs

	c.JSON(http.StatusOK, res)
}

func main() {
	env.InitGlobalEnv()
	InitializeQueues()
	serveStorage = kv.NewKV("serve")

	s, _ := env.Env.GetUserService("serve")
	static_files_path := s.GetParamString("static_files_path")
	external_host := s.GetParamString("external_host")
	external_port := s.GetParamInt64("external_port")

	log.Println("environment initialized")

	zap.L().Info("serve environment",
		zap.String("static_files_path", static_files_path),
		zap.String("external_host", external_host),
		zap.Int64("external_port", external_port),
	)

	/////////////////////
	// Server/API
	engine := gin.Default()
	engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/search/start")
	})
	engine.GET("/search", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/search/start")
	})
	engine.StaticFS("/static", gin.Dir(static_files_path, true))
	engine.GET("/search/:host", ServeHost)
	v1 := engine.Group("/api")
	{
		v1.GET("/heartbeat", common.Heartbeat)
		v1.POST("/search", SearchHandler)
		v1.GET("/databases", DatabasesHandler)
		v1.GET("/stats", MultiStatsHandler)
	}

	//engine.Use(static.Serve("/static", static.LocalFile(static_files_path, true)))

	// Serve up the search page
	// fs := http.FileServer(http.Dir(static_files_path))
	// engine.Handle("/static/*", http.StripPrefix("/static/", fs))
	// engine.Static("/static", static_files_path)

	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, engine)

}

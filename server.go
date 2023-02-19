package main

import (
	"embed"
	"go_sync/ws"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed frontend/dist/*
var FS embed.FS

func StartGinServer() {

	hub := ws.NewHub()
	go hub.Run()

	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"msg": "ok",
		})
	})
	staticFiles, _ := fs.Sub(FS, "frontend/dist")
	r.StaticFS("/static", http.FS(staticFiles))

	r.MaxMultipartMemory = 200 << 20

	r.POST("/api/v1/texts", TextsController)
	r.GET("/api/v1/addresses", AddressesController)
	r.GET("/uploads/:path", UploadsController)
	r.GET("/api/v1/qrcodes", QrcodesController)
	r.POST("/api/v1/files", FilesController)

	r.GET("/ws", func(ctx *gin.Context) {
		ws.HttpController(ctx, hub)
	})

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/static/") {
			reader, err := staticFiles.Open("index.html")
			if err != nil {
				log.Fatal(err)
			}
			defer reader.Close()
			stat, err := reader.Stat()
			if err != nil {
				log.Fatal(err)
			}
			c.DataFromReader(http.StatusOK, stat.Size(), "text/html;charset=utf-8", reader, nil)
		} else {
			c.Status(http.StatusNotFound)
		}
	})

	r.Run(":" + port)
}

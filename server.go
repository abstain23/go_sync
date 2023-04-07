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

func CatchPanic() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic: %v\n", r)

				ctx.JSON(http.StatusInternalServerError, gin.H{})

				ctx.Abort()
			}

		}()

		ctx.Next()
	}
}

func StartGinServer() {

	hub := ws.NewHub()
	go hub.Run()

	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	r.Use(CatchPanic())

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

	r.GET("/ws_ping", func(ctx *gin.Context) {
		ws.HttpController2(ctx, chChromeDie, chBackendDie)
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

package main

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/zserge/lorca"
)

//go:embed frontend/dist/*
var FS embed.FS

func main() {
	go func() {
		gin.SetMode(gin.DebugMode)
		r := gin.Default()
		r.GET("/", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{
				"msg": "ok",
			})
		})
		staticFils, _ := fs.Sub(FS, "frontend/dist")
		r.StaticFS("/static", http.FS(staticFils))
		r.Run(":8082")
	}()

	ui, _ := lorca.New("http://127.0.0.1:8082/static/index.html", "", 800, 600, "--disable-sync", "--disable-translate", "--disable-automation")
	chSignal := make(chan os.Signal, 1)

	signal.Notify(chSignal, os.Interrupt)

	select {
	case <-ui.Done():
	case <-chSignal:
	}
	ui.Close()
}

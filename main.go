package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"github.com/zserge/lorca"
)

//go:embed frontend/dist/*
var FS embed.FS

func main() {
	port := "27149"
	go func() {
		gin.SetMode(gin.DebugMode)
		r := gin.Default()
		r.GET("/", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{
				"msg": "ok",
			})
		})
		staticFiles, _ := fs.Sub(FS, "frontend/dist")
		r.StaticFS("/static", http.FS(staticFiles))

		r.POST("/api/v1/texts", TextsController)
		r.GET("/api/v1/addresses", AddressesController)
		r.GET("/uploads/:path", UploadsController)
		r.GET("/api/v1/qrcodes", QrcodesController)
		r.POST("/api/v1/files", FilesController)

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
	}()

	ui, _ := lorca.New("http://127.0.0.1:"+port+"/static/index.html", "", 800, 600, "--disable-sync", "--disable-translate", "--disable-automation")
	chSignal := make(chan os.Signal, 1)

	signal.Notify(chSignal, os.Interrupt)

	select {
	case <-ui.Done():
	case <-chSignal:
	}
	ui.Close()
}

var uploadsDir string

func init() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exeDir := filepath.Dir(exe)
	uploadsDir = filepath.Join(exeDir, "uploads")
	if !Exists(uploadsDir) {
		fmt.Println("创建文件夹")
		os.Mkdir(uploadsDir, os.ModePerm)
	}
}

// 判断所给路径文件或者文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	}

	return true
}

func FilesController(c *gin.Context) {
	file, err := c.FormFile("raw")
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}
	fileName := time.Now().Format("2006-01-02_15_04_05")

	if err != nil {
		log.Fatal(err)
	}
	fullpath := filepath.Join(fileName + filepath.Ext(file.Filename))
	fileErr := c.SaveUploadedFile(file, filepath.Join(uploadsDir, fullpath))
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + fullpath})
}

func QrcodesController(c *gin.Context) {
	if content := c.Query("content"); content != "" {
		png, err := qrcode.Encode(content, qrcode.Medium, 256)
		if err != nil {
			log.Fatal(err)
		}
		c.Data(http.StatusOK, "image/png", png)
	} else {
		c.Status(http.StatusBadRequest)
	}
}

func UploadsController(c *gin.Context) {
	if path := c.Param("path"); path != "" {
		target := filepath.Join(uploadsDir, path)
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename="+path)
		c.Header("Content-Type", "application/octet-stream")
		c.File(target)
	} else {
		c.Status(http.StatusNotFound)
	}
}

func AddressesController(c *gin.Context) {
	addrs, _ := net.InterfaceAddrs()
	var result []string
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				result = append(result, ipnet.IP.String())
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"addresses": result})
}

func TextsController(c *gin.Context) {
	var json struct {
		Raw string `json:"raw"`
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		fileName := time.Now().Format("2006-01-02_15_04_05") + ".txt"
		err = os.WriteFile(filepath.Join(uploadsDir, fileName), []byte(json.Raw), 0644) // 将 json.Raw 写入文件
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{"url": "/" + "uploads/" + fileName}) // 返回文件的绝对路径（不含 exe 所在目录）
	}

}

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

func FilesController(c *gin.Context) {
	file, err := c.FormFile("raw")
	if err != nil {
		log.Fatal("err1: ", err)
	}

	fileName := time.Now().Format("2006-01-02_15_04_05")

	if err != nil {
		log.Fatal(err)
	}
	fullpath := filepath.Join(fileName + filepath.Ext(file.Filename))
	fileErr := c.SaveUploadedFile(file, filepath.Join(uploadsDir, fullpath))
	if fileErr != nil {
		log.Fatal("fileErr", fileErr)
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

func BigFileController(c *gin.Context) {
	fmt.Println("ffff")
	c.Request.ParseMultipartForm(32 << 20)
	file, _, err := c.Request.FormFile("row")
	if err != nil {
		fmt.Fprintln(c.Writer, err)
		return
	}
	defer file.Close()
	f, err := os.OpenFile("./test.txt", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Fprintf(c.Writer, "Failed to open the file for writing")
		return
	}
	defer f.Close()
	io.Copy(f, file)
	fmt.Fprintln(c.Writer, "File uploaded successfully")
	c.JSON(http.StatusOK, gin.H{"url": "/" + "uploads/" + "test.txt"})
}

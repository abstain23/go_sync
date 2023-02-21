package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/zserge/lorca"
)

var port = "27149"

func main() {

	go StartGinServer()

	ui, _ := lorca.New("http://127.0.0.1:"+port+"/static/index.html", "", 800, 600)
	chSignal := make(chan os.Signal, 1)

	signal.Notify(chSignal, os.Interrupt)

	select {
	case <-ui.Done():
		fmt.Println("ui.Done")
	case <-chSignal:
		fmt.Println("signal, notify")
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

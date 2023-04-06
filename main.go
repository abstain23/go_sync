package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
)

var port = "27149"

var defaultChromeArgs = []string{
	"--disable-background-networking",
	"--disable-background-timer-throttling",
	"--disable-backgrounding-occluded-windows",
	"--disable-breakpad",
	"--disable-client-side-phishing-detection",
	"--disable-default-apps",
	"--disable-dev-shm-usage",
	"--disable-infobars",
	"--disable-extensions",
	"--disable-features=site-per-process",
	"--disable-hang-monitor",
	"--disable-ipc-flooding-protection",
	"--disable-popup-blocking",
	"--disable-prompt-on-repost",
	"--disable-renderer-backgrounding",
	// "--disable-sync",
	"--disable-translate",
	"--disable-windows10-custom-titlebar",
	"--metrics-recording-only",
	"--no-first-run",
	"--no-default-browser-check",
	"--safebrowsing-disable-auto-update",
	// "--enable-automation",
	"--password-store=basic",
	"--use-mock-keychain",
}

// go build -ldflags -H=windowsgui .

func LocateChrome() string {

	// If env variable "LORCACHROME" specified and it exists
	if path, ok := os.LookupEnv("LORCACHROME"); ok {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	var paths []string
	switch runtime.GOOS {
	case "darwin":
		paths = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
		}
	case "windows":
		paths = []string{
			os.Getenv("LocalAppData") + "/Google/Chrome/Application/chrome.exe",
			os.Getenv("ProgramFiles") + "/Google/Chrome/Application/chrome.exe",
			os.Getenv("ProgramFiles(x86)") + "/Google/Chrome/Application/chrome.exe",
			os.Getenv("LocalAppData") + "/Chromium/Application/chrome.exe",
			os.Getenv("ProgramFiles") + "/Chromium/Application/chrome.exe",
			os.Getenv("ProgramFiles(x86)") + "/Chromium/Application/chrome.exe",
			os.Getenv("ProgramFiles(x86)") + "/Microsoft/Edge/Application/msedge.exe",
			os.Getenv("ProgramFiles") + "/Microsoft/Edge/Application/msedge.exe",
		}
	default:
		paths = []string{
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		return path
	}
	return ""
}

func startBrowser(chChromeDie chan struct{}, chBackendDie chan struct{}) {
	chromePath := LocateChrome()
	name, err := ioutil.TempDir("", "lorca")
	if err != nil {
		panic(err)
	}
	dir := name
	args := append(defaultChromeArgs, fmt.Sprintf("--app=%s", "http://127.0.0.1:"+port+"/static/index.html"))
	args = append(args, fmt.Sprintf("--user-data-dir=%s", dir))
	args = append(args, fmt.Sprintf("--window-size=%d,%d", 800, 600))
	args = append(args, "--remote-debugging-port=0")
	cmd := exec.Command(chromePath, args...)
	cmd.Start()
	go func() {
		<-chBackendDie
		cmd.Process.Kill()
	}()
	go func() {
		cmd.Wait()
		chChromeDie <- struct{}{}
	}()
}

func main() {

	chChromeDie := make(chan struct{})
	chBackendDie := make(chan struct{})
	chSignal := listenToInterrupt()

	go StartGinServer()

	go startBrowser(chChromeDie, chBackendDie)

	for {
		select {
		case <-chSignal:
			chBackendDie <- struct{}{}
		case <-chChromeDie:
			os.Exit(0)
		}
	}

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

func listenToInterrupt() chan os.Signal {
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, os.Interrupt)
	return chSignal
}

// 判断所给路径文件或者文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	}

	return true
}

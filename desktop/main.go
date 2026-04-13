package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	winoptions "github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logFile := initDesktopLogging()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("fatal: panic: %v", r)
		}
		if logFile != nil {
			_ = logFile.Close()
		}
	}()

	app := NewApp()
	err := wails.Run(&options.App{
		Title:            "K8V",
		Width:            1240,
		Height:           720,
		MinWidth:         960,
		MinHeight:        620,
		Frameless:        false,
		Fullscreen:       false,
		DisableResize:    false,
		WindowStartState: options.Normal,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: options.NewRGB(13, 17, 23),
		Windows:          desktopWindowsOptions(),
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Printf("fatal: wails run error: %v", err)
		println("Error:", err.Error())
	}
}

func initDesktopLogging() *os.File {
	configRoot, err := os.UserConfigDir()
	if err != nil || configRoot == "" {
		return nil
	}
	logDir := filepath.Join(configRoot, "k8s-recovery-visualizer", "logs")
	if mkErr := os.MkdirAll(logDir, 0o700); mkErr != nil {
		return nil
	}
	logPath := filepath.Join(logDir, "k8v-startup.log")
	f, openErr := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if openErr != nil {
		return nil
	}
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	log.Printf("startup: %s", time.Now().UTC().Format(time.RFC3339))
	log.Printf("runtime: goos=%s goarch=%s", runtime.GOOS, runtime.GOARCH)
	log.Printf("configDir: %s", configRoot)
	if runtime.GOOS == "windows" {
		webviewPath := filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "EdgeWebView", "Application", "msedgewebview2.exe")
		if webviewPath == "" {
			webviewPath = filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "EdgeWebView", "Application", "msedgewebview2.exe")
		}
		if webviewPath != "" {
			if _, statErr := os.Stat(webviewPath); statErr != nil {
				log.Printf("webview2: not found at %s (%v)", webviewPath, statErr)
			} else {
				log.Printf("webview2: found at %s", webviewPath)
			}
		}
	}
	return f
}

func desktopWindowsOptions() *winoptions.Options {
	configRoot, err := os.UserConfigDir()
	if err != nil || configRoot == "" {
		return &winoptions.Options{}
	}

	return &winoptions.Options{
		Theme:               winoptions.Dark,
		DisableWindowIcon:   false,
		WindowClassName:     "K8VMainWindow",
		WebviewUserDataPath: filepath.Join(configRoot, "k8s-recovery-visualizer", "webview2"),
	}
}

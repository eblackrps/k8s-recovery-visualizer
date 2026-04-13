package main

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	winoptions "github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
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
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
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

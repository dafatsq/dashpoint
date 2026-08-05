package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var frontendAssets embed.FS

func main() {
	err := wails.Run(&options.App{
		Title:  "DashPoint POS",
		Width:  1440,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: frontendAssets,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

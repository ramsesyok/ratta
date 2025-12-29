package main

import (
	"embed"
	"flag"
	"os"

	"ratta/internal/app/contractorinit"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if handled, code := runCLI(); handled {
		os.Exit(code)
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "ratta",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func runCLI() (bool, int) {
	if len(os.Args) < 2 {
		return false, 0
	}
	if os.Args[1] != "init" || len(os.Args) < 3 || os.Args[2] != "contractor" {
		return false, 0
	}

	fs := flag.NewFlagSet("init contractor", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite existing contractor.json")
	if err := fs.Parse(os.Args[3:]); err != nil {
		return true, 1
	}

	exePath, err := os.Executable()
	if err != nil {
		return true, 1
	}
	if err := contractorinit.Run(exePath, *force, contractorinit.ConsolePrompter{}); err != nil {
		return true, 1
	}
	return true, 0
}

// main.go はアプリ起動とCLI初期化を担い、UI詳細は扱わない。
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

// main は Wails アプリとCLIモードの起動を行う。
// 目的: CLI 初期化とGUI起動を切り替える。
// 入力: コマンドライン引数。
// 出力: なし。
// エラー: CLI 処理の失敗時は終了コードで示す。
// 副作用: プロセス終了やアプリ起動を行う。
// 並行性: 単一ゴルーチンで実行する。
// 不変条件: CLI が処理された場合は GUI を起動しない。
// 関連DD: DD-BE-002, DD-CLI-002
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

// runCLI は CLI モードの初期化コマンドを処理する。
// 目的: init contractor を検出し認証ファイル生成を実行する。
// 入力: os.Args の内容。
// 出力: handled は CLI を処理したか、code は終了コード。
// エラー: 失敗時は handled=true と code=1 を返す。
// 副作用: contractor.json 生成やプロセス終了コードに影響する。
// 並行性: 単一ゴルーチンで実行する。
// 不変条件: 対象外の引数は handled=false を返す。
// 関連DD: DD-CLI-002, DD-CLI-003, DD-CLI-004
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
	if runErr := contractorinit.Run(exePath, *force, contractorinit.ConsolePrompter{}); runErr != nil {
		return true, 1
	}
	return true, 0
}

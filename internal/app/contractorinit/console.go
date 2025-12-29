// console.go は CLI の入力実装を担い、認証ファイル生成の制御は扱わない。
package contractorinit

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// ConsolePrompter は DD-CLI-003 の端末入力を担当する。
type ConsolePrompter struct{}

// PromptHidden は端末に表示せずパスワード入力を受け付ける。
// 目的: 画面に表示せず安全にパスワード文字列を取得する。
// 入力: label は入力プロンプト文字列。
// 出力: 入力された文字列とエラー。
// エラー: 端末入力に失敗した場合に返す。
// 副作用: 標準出力にプロンプトと改行を出力する。
// 並行性: 同時入力は想定しない。
// 不変条件: 入力内容は表示されない。
// 関連DD: DD-CLI-003
func (c ConsolePrompter) PromptHidden(label string) (string, error) {
	fmt.Print(label)
	input, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return string(input), nil
}

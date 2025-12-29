package contractorinit

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// ConsolePrompter は DD-CLI-003 の端末入力を担当する。
type ConsolePrompter struct{}

// PromptHidden は端末に表示せずパスワード入力を受け付ける。
func (c ConsolePrompter) PromptHidden(label string) (string, error) {
	fmt.Print(label)
	input, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(input), nil
}

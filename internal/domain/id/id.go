// Package id は各種ID生成を担い、IDの利用方法や保存先は扱わない。
// 乱数の供給元は差し替え可能とする。
package id

import (
	"fmt"

	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	nanoAlphabet = "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	nanoIDLength = 9
)

var (
	uuidV7Generator = uuid.NewV7
	nanoidGenerate  = gonanoid.Generate
)

// NewIssueID は DD-DATA-003 の issue_id 仕様に従い nanoid (9 文字) を生成する。
func NewIssueID() (string, error) {
	return newNanoID()
}

// NewAttachmentID は DD-DATA-005 の attachment_id 仕様に従い nanoid (9 文字) を生成する。
func NewAttachmentID() (string, error) {
	return newNanoID()
}

// NewCommentID は DD-DATA-004 の comment_id 仕様に従い UUID v7 を生成する。
func NewCommentID() (string, error) {
	value, err := uuidV7Generator()
	if err != nil {
		return "", fmt.Errorf("uuid v7: %w", err)
	}
	return value.String(), nil
}

// newNanoID は DD-DATA-003/DD-DATA-005 の ID 仕様に従い nanoid (9 文字) を生成する。
func newNanoID() (string, error) {
	value, err := nanoidGenerate(nanoAlphabet, nanoIDLength)
	if err != nil {
		return "", fmt.Errorf("nanoid: %w", err)
	}
	return value, nil
}

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

var uuidV7Generator = uuid.NewV7
var nanoidGenerate = gonanoid.Generate

// NewIssueID は nanoid (9 文字) の issue_id を生成する。
func NewIssueID() (string, error) {
	return newNanoID()
}

// NewAttachmentID は nanoid (9 文字) の attachment_id を生成する。
func NewAttachmentID() (string, error) {
	return newNanoID()
}

// NewCommentID は UUID v7 の comment_id を生成する。
func NewCommentID() (string, error) {
	value, err := uuidV7Generator()
	if err != nil {
		return "", fmt.Errorf("uuid v7: %w", err)
	}
	return value.String(), nil
}

func newNanoID() (string, error) {
	value, err := nanoidGenerate(nanoAlphabet, nanoIDLength)
	if err != nil {
		return "", fmt.Errorf("nanoid: %w", err)
	}
	return value, nil
}

package attachmentstore

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"ratta/internal/domain/id"
)

const (
	maxFileNameLength = 255
	attachmentDirExt  = ".files"
)

var now = time.Now
var newAttachmentID = id.NewAttachmentID
var renameFile = os.Rename
var removeFile = os.Remove
var createTempFile = func(dir, base string) (io.WriteCloser, string, error) {
	timestamp := now().Unix()
	tmpName := fmt.Sprintf("%s.tmp.%d.%d", base, os.Getpid(), timestamp)
	tmpPath := filepath.Join(dir, tmpName)
	file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, "", err
	}
	return file, tmpPath, nil
}

// Input は DD-DATA-005 の添付情報をもとに保存対象を表す。
type Input struct {
	OriginalName string
	Data         []byte
}

// SavedAttachment は DD-DATA-005 の添付保存結果を表す。
type SavedAttachment struct {
	AttachmentID string
	OriginalName string
	StoredName   string
	RelativePath string
	FullPath     string
}

// SaveAll は DD-DATA-005 の格納ルールに従い、添付ファイルを保存する。
// 保存後に JSON 更新が失敗した場合は、返却する rollback を呼び出して削除する。
func SaveAll(issueDir, issueID string, inputs []Input) ([]SavedAttachment, func() error, error) {
	if len(inputs) == 0 {
		return nil, func() error { return nil }, nil
	}

	attachDir := filepath.Join(issueDir, issueID+attachmentDirExt)
	if err := os.MkdirAll(attachDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create attachment dir: %w", err)
	}

	var saved []SavedAttachment
	for _, input := range inputs {
		record, err := saveOne(attachDir, issueID, input)
		if err != nil {
			_ = removeAll(saved)
			return nil, nil, err
		}
		saved = append(saved, record)
	}

	return saved, func() error { return removeAll(saved) }, nil
}

func saveOne(attachDir, issueID string, input Input) (SavedAttachment, error) {
	attachmentID, err := newAttachmentID()
	if err != nil {
		return SavedAttachment{}, fmt.Errorf("generate attachment id: %w", err)
	}

	sanitized := sanitizeFileName(input.OriginalName)
	storedName, err := buildStoredName(attachDir, attachmentID, sanitized)
	if err != nil {
		return SavedAttachment{}, err
	}

	fullPath := filepath.Join(attachDir, storedName)
	if err := writeWithTemp(attachDir, storedName, input.Data); err != nil {
		return SavedAttachment{}, err
	}

	return SavedAttachment{
		AttachmentID: attachmentID,
		OriginalName: input.OriginalName,
		StoredName:   storedName,
		RelativePath: fmt.Sprintf("%s%s/%s", issueID, attachmentDirExt, storedName),
		FullPath:     fullPath,
	}, nil
}

// writeWithTemp は DD-PERSIST-002 を参考に、一時ファイル経由で保存する。
func writeWithTemp(dir, base string, data []byte) error {
	writer, tmpPath, err := createTempFile(dir, base)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		_ = removeFile(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := writer.Close(); err != nil {
		_ = removeFile(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := renameFile(tmpPath, filepath.Join(dir, base)); err != nil {
		_ = removeFile(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

// removeAll は DD-DATA-005 のロールバック要件に従い保存済み添付を削除する。
func removeAll(saved []SavedAttachment) error {
	var errs []string
	for _, record := range saved {
		if err := removeFile(record.FullPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("remove attachments: %s", strings.Join(errs, ", "))
	}
	return nil
}

// buildStoredName は DD-DATA-005 の stored_name 仕様に従い衝突回避名を作る。
func buildStoredName(dir, attachmentID, sanitizedName string) (string, error) {
	namePart, ext := splitExt(sanitizedName)
	basePrefix := attachmentID + "_"
	namePart = trimToLength(namePart, maxFileNameLength-utf8.RuneCountInString(basePrefix)-utf8.RuneCountInString(ext))
	if namePart == "" {
		namePart = "_"
	}

	base := basePrefix + namePart
	candidate := base + ext
	if !exists(filepath.Join(dir, candidate)) {
		return candidate, nil
	}

	for i := 1; i < 1000; i++ {
		suffix := "_" + strconv.Itoa(i)
		limit := maxFileNameLength - utf8.RuneCountInString(basePrefix) - utf8.RuneCountInString(ext) - utf8.RuneCountInString(suffix)
		trimmed := trimToLength(namePart, limit)
		if trimmed == "" {
			trimmed = "_"
		}
		candidate = basePrefix + trimmed + suffix + ext
		if !exists(filepath.Join(dir, candidate)) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("stored name collision limit reached")
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// sanitizeFileName は DD-DATA-005 の Windows 禁止文字ルールに従って整形する。
func sanitizeFileName(name string) string {
	if name == "" {
		return "_"
	}

	replacer := func(r rune) rune {
		switch r {
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
			return '_'
		default:
			return r
		}
	}
	cleaned := strings.Map(replacer, name)

	runes := []rune(cleaned)
	for len(runes) > 0 {
		last := runes[len(runes)-1]
		if last != '.' && last != ' ' {
			break
		}
		runes[len(runes)-1] = '_'
	}
	cleaned = string(runes)
	cleaned = trimToLength(cleaned, maxFileNameLength)
	if cleaned == "" {
		return "_"
	}
	return cleaned
}

// trimToLength は DD-DATA-005 の 255 文字制限に合わせて切り詰める。
func trimToLength(value string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

// splitExt は拡張子を分離し、DD-DATA-005 の名称組み立てに使う。
func splitExt(name string) (string, string) {
	ext := filepath.Ext(name)
	if ext == "" || ext == name {
		return name, ""
	}
	return strings.TrimSuffix(name, ext), ext
}

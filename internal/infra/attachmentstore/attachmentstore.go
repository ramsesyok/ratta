// Package attachmentstore は添付ファイルの保存を担い、UI表示や課題更新は扱わない。
// ファイル保存の一時ファイル処理はこのパッケージ内で完結させる。
package attachmentstore

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"ratta/internal/domain/id"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxFileNameLength = 255
	attachmentDirExt  = ".files"
)

var (
	now             = time.Now
	newAttachmentID = id.NewAttachmentID
	renameFile      = os.Rename
	removeFile      = os.Remove
	createTempFile  = func(dir, base string) (io.WriteCloser, string, error) {
		timestamp := now().Unix()
		tmpName := fmt.Sprintf("%s.tmp.%d.%d", base, os.Getpid(), timestamp)
		tmpPath := filepath.Join(dir, tmpName)
		// #nosec G304 -- 添付保存ディレクトリ配下の一時ファイルのみを作成するため安全。
		file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			return nil, "", fmt.Errorf("open temp file: %w", err)
		}
		return file, tmpPath, nil
	}
)

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
// 目的: 複数添付を保存し、ロールバック関数を返却する。
// 入力: issueDir は課題ディレクトリ、issueID は課題ID、inputs は添付入力群。
// 出力: 保存済み添付一覧、ロールバック関数、エラー。
// エラー: 保存失敗やロールバック失敗時に返す。
// 副作用: 添付ディレクトリ作成とファイル書き込みを行う。
// 並行性: 同一課題への同時保存は想定しない。
// 不変条件: 保存に失敗した場合は保存済み添付を削除する。
// 関連DD: DD-DATA-005
func SaveAll(issueDir, issueID string, inputs []Input) ([]SavedAttachment, func() error, error) {
	if len(inputs) == 0 {
		return nil, func() error { return nil }, nil
	}

	attachDir := filepath.Join(issueDir, issueID+attachmentDirExt)
	if err := os.MkdirAll(attachDir, 0o750); err != nil {
		return nil, nil, fmt.Errorf("create attachment dir: %w", err)
	}

	saved := make([]SavedAttachment, 0, len(inputs))
	for _, input := range inputs {
		record, err := saveOne(attachDir, issueID, input)
		if err != nil {
			if cleanupErr := removeAll(saved); cleanupErr != nil {
				return nil, nil, fmt.Errorf("cleanup attachments failed: %w; cleanup error: %s", err, cleanupErr.Error())
			}
			return nil, nil, err
		}
		saved = append(saved, record)
	}

	return saved, func() error { return removeAll(saved) }, nil
}

// saveOne は DD-DATA-005 の保存単位で添付を1件保存する。
// 目的: 添付IDを発行しファイル名を正規化して保存する。
// 入力: attachDir は保存先、issueID は課題ID、input は添付入力。
// 出力: SavedAttachment とエラー。
// エラー: ID生成や保存失敗時に返す。
// 副作用: ファイルを作成する。
// 並行性: 同一ディレクトリへの同時保存は想定しない。
// 不変条件: StoredName は sanitize と衝突回避に従う。
// 関連DD: DD-DATA-005
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
	if writeErr := writeWithTemp(attachDir, storedName, input.Data); writeErr != nil {
		return SavedAttachment{}, writeErr
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
// 目的: 原子的に添付ファイルを書き込む。
// 入力: dir は保存先、base はファイル名、data は内容。
// 出力: 成功時は nil、失敗時はエラー。
// エラー: 一時ファイル作成、書き込み、リネーム失敗時に返す。
// 副作用: 一時ファイル作成・削除とファイル更新を行う。
// 並行性: 同一ファイルへの同時書き込みは想定しない。
// 不変条件: 書き込み失敗時は目的ファイルを更新しない。
// 関連DD: DD-PERSIST-002
func writeWithTemp(dir, base string, data []byte) error {
	writer, tmpPath, err := createTempFile(dir, base)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	if _, writeErr := writer.Write(data); writeErr != nil {
		closeErr := writer.Close()
		removeErr := removeFile(tmpPath)
		if closeErr != nil {
			return fmt.Errorf("write temp file failed: %w; close error: %s", writeErr, closeErr.Error())
		}
		if removeErr != nil {
			return fmt.Errorf("write temp file failed: %w; cleanup error: %s", writeErr, removeErr.Error())
		}
		return fmt.Errorf("write temp file: %w", writeErr)
	}

	if closeErr := writer.Close(); closeErr != nil {
		removeErr := removeFile(tmpPath)
		if removeErr != nil {
			return fmt.Errorf("close temp file failed: %w; cleanup error: %s", closeErr, removeErr.Error())
		}
		return fmt.Errorf("close temp file: %w", closeErr)
	}

	if renameErr := renameFile(tmpPath, filepath.Join(dir, base)); renameErr != nil {
		removeErr := removeFile(tmpPath)
		if removeErr != nil {
			return fmt.Errorf("rename temp file failed: %w; cleanup error: %s", renameErr, removeErr.Error())
		}
		return fmt.Errorf("rename temp file: %w", renameErr)
	}

	return nil
}

// removeAll は DD-DATA-005 のロールバック要件に従い保存済み添付を削除する。
// 目的: 保存済み添付を一括削除する。
// 入力: saved は保存済み添付の一覧。
// 出力: 成功時は nil、失敗時はエラー。
// エラー: 削除に失敗したファイルがある場合に返す。
// 副作用: 添付ファイルを削除する。
// 並行性: 同時削除は想定しない。
// 不変条件: エラー時は削除できなかったパスを集約する。
// 関連DD: DD-DATA-005
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

	return "", errors.New("stored name collision limit reached")
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
// trimToLength は DD-DATA-005 の 255 文字制限に合わせて切り詰める。
// 目的: 文字数制限に合わせて末尾を切り詰める。
// 入力: value は対象文字列、maxLen は許容文字数。
// 出力: 切り詰め後の文字列。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 返却値の長さは maxLen 以下。
// 関連DD: DD-DATA-005
func trimToLength(value string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= maxLen {
		return value
	}
	return string(runes[:maxLen])
}

// splitExt は拡張子を分離し、DD-DATA-005 の名称組み立てに使う。
func splitExt(name string) (string, string) {
	ext := filepath.Ext(name)
	if ext == "" || ext == name {
		return name, ""
	}
	return strings.TrimSuffix(name, ext), ext
}

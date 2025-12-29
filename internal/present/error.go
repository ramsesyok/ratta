// Package present は UI へ返すDTOとエラー表現を提供し、ドメイン実装は扱わない。
// エラー分類ロジックはここで完結させる。
package present

import (
	"errors"
	"ratta/internal/domain/issue"
	"strings"
)

const (
	ErrorValidation = "E_VALIDATION"
	ErrorPermission = "E_PERMISSION"
	ErrorNotFound   = "E_NOT_FOUND"
	ErrorConflict   = "E_CONFLICT"
	ErrorCrypto     = "E_CRYPTO"
	ErrorInternal   = "E_INTERNAL"
)

// Ok は DD-BE-003 の成功レスポンスを作る。
func Ok(data any) Response {
	return Response{Ok: true, Data: data}
}

// Fail は DD-BE-003 の失敗レスポンスを作る。
func Fail(err error) Response {
	return Response{Ok: false, Error: MapError(err)}
}

// MapError は DD-BE-003 の APIErrorDTO へ変換する。
// 目的: 内部エラーをUI向けの共通エラー形式に正規化する。
// 入力: err は内部エラー。
// 出力: APIErrorDTO へのポインタ。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: err が nil の場合は nil を返す。
// 関連DD: DD-BE-003
func MapError(err error) *APIErrorDTO {
	if err == nil {
		return nil
	}

	var validationErrors issue.ValidationErrors
	if errors.As(err, &validationErrors) {
		return &APIErrorDTO{
			ErrorCode: ErrorValidation,
			Message:   "Validation failed.",
			Detail:    err.Error(),
		}
	}
	var validationError *issue.ValidationError
	if errors.As(err, &validationError) {
		return &APIErrorDTO{
			ErrorCode: ErrorValidation,
			Message:   "Validation failed.",
			Detail:    err.Error(),
		}
	}

	message := err.Error()
	code := classifyError(message)
	return &APIErrorDTO{
		ErrorCode: code,
		Message:   message,
	}
}

func classifyError(message string) string {
	switch {
	case strings.Contains(message, "permission"):
		return ErrorPermission
	case strings.Contains(message, "not found"):
		return ErrorNotFound
	case strings.Contains(message, "conflict"),
		strings.Contains(message, "read-only"),
		strings.Contains(message, "schema invalid"),
		strings.Contains(message, "not empty"):
		return ErrorConflict
	case strings.Contains(message, "password verification failed"),
		strings.Contains(message, "crypto"):
		return ErrorCrypto
	default:
		return ErrorInternal
	}
}

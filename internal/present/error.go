package present

import (
	"errors"
	"strings"

	"ratta/internal/domain/issue"
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

// MapError は DD-BE-003 の ApiErrorDTO へ変換する。
func MapError(err error) *ApiErrorDTO {
	if err == nil {
		return nil
	}

	var validationErrors issue.ValidationErrors
	if errors.As(err, &validationErrors) {
		return &ApiErrorDTO{
			ErrorCode: ErrorValidation,
			Message:   "Validation failed.",
			Detail:    err.Error(),
		}
	}
	var validationError *issue.ValidationError
	if errors.As(err, &validationError) {
		return &ApiErrorDTO{
			ErrorCode: ErrorValidation,
			Message:   "Validation failed.",
			Detail:    err.Error(),
		}
	}

	message := err.Error()
	code := classifyError(message)
	return &ApiErrorDTO{
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

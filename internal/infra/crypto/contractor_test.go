package crypto

import (
	"bytes"
	"errors"
	"testing"
)

func TestGenerateAndVerifyContractorAuth(t *testing.T) {
	// DD-CLI-005 の手順で生成したデータが同じパスワードで検証できることを確認する。
	previousReader := randReader
	randReader = bytes.NewReader(bytes.Repeat([]byte{0x01}, saltSizeBytes+nonceSizeBytes))
	t.Cleanup(func() { randReader = previousReader })

	auth, err := GenerateContractorAuth("secret")
	if err != nil {
		t.Fatalf("GenerateContractorAuth error: %v", err)
	}
	if auth.FormatVersion != formatVersion {
		t.Fatalf("unexpected format version: %d", auth.FormatVersion)
	}
	if auth.KDF != kdfName {
		t.Fatalf("unexpected kdf: %s", auth.KDF)
	}
	if auth.KDFIterations != kdfIterations {
		t.Fatalf("unexpected iterations: %d", auth.KDFIterations)
	}

	ok, err := VerifyPassword(auth, "secret")
	if err != nil {
		t.Fatalf("VerifyPassword error: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify")
	}
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	// パスワードが異なる場合は検証に失敗することを確認する。
	previousReader := randReader
	randReader = bytes.NewReader(bytes.Repeat([]byte{0x02}, saltSizeBytes+nonceSizeBytes))
	t.Cleanup(func() { randReader = previousReader })

	auth, err := GenerateContractorAuth("secret")
	if err != nil {
		t.Fatalf("GenerateContractorAuth error: %v", err)
	}

	ok, err := VerifyPassword(auth, "wrong")
	if err == nil {
		t.Fatal("expected password mismatch error")
	}
	if !errors.Is(err, ErrPasswordMismatch) {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected password verification to fail")
	}
}

func TestVerifyPassword_UnsupportedKDF(t *testing.T) {
	// 未対応のKDF設定がエラーになることを確認する。
	auth := ContractorAuth{
		KDF:           "unknown",
		KDFIterations: kdfIterations,
		SaltB64:       "AA==",
		NonceB64:      "AA==",
		CiphertextB64: "AA==",
	}
	if _, err := VerifyPassword(auth, "secret"); !errors.Is(err, ErrUnsupportedKDF) {
		t.Fatalf("expected unsupported kdf error, got: %v", err)
	}
}

func TestSetRandReader_Restore(t *testing.T) {
	// SetRandReader が元の乱数リーダーを復元することを確認する。
	previous := randReader
	restore := SetRandReader(bytes.NewReader([]byte{0x01}))
	restore()
	if randReader != previous {
		t.Fatal("expected randReader to be restored")
	}
}

func TestEncryptFixed_InvalidKey(t *testing.T) {
	// 不正な鍵長で暗号化に失敗することを確認する。
	if _, err := encryptFixed([]byte("short"), []byte("0123456789ABCDEF")); err == nil {
		t.Fatal("expected encrypt error")
	}
}

func TestGenerateContractorAuth_EmptyPassword(t *testing.T) {
	// 空パスワードが拒否されることを確認する。
	if _, err := GenerateContractorAuth(""); err == nil {
		t.Fatal("expected password required error")
	}
}

func TestGenerateContractorAuth_RandFailure(t *testing.T) {
	// 乱数読み取り失敗時にエラーとなることを確認する。
	previous := randReader
	randReader = bytes.NewReader([]byte{0x01})
	t.Cleanup(func() { randReader = previous })

	if _, err := GenerateContractorAuth("secret"); err == nil {
		t.Fatal("expected rand read error")
	}
}

func TestVerifyPassword_DecodeError(t *testing.T) {
	// Base64 変換に失敗した場合にエラーとなることを確認する。
	auth := ContractorAuth{
		KDF:           kdfName,
		KDFIterations: kdfIterations,
		SaltB64:       "%%%invalid",
		NonceB64:      "AA==",
		CiphertextB64: "AA==",
	}
	if _, err := VerifyPassword(auth, "secret"); err == nil {
		t.Fatal("expected decode error")
	}
}

package crypto

import (
	"bytes"
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
	if err != nil {
		t.Fatalf("VerifyPassword error: %v", err)
	}
	if ok {
		t.Fatal("expected password verification to fail")
	}
}

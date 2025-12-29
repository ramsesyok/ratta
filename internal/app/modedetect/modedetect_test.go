package modedetect

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"ratta/internal/domain/mode"
	"ratta/internal/infra/crypto"
	"ratta/internal/infra/jsonfmt"
	"ratta/internal/infra/schema"
)

func TestDetectMode_NoAuthFile(t *testing.T) {
	// auth/contractor.json が無ければ Vendor でパスワード不要になることを確認する。
	dir := t.TempDir()
	service := NewService(filepath.Join(dir, "ratta.exe"), nil)

	gotMode, requiresPassword, err := service.DetectMode()
	if err != nil {
		t.Fatalf("DetectMode error: %v", err)
	}
	if gotMode != mode.ModeVendor {
		t.Fatalf("unexpected mode: %s", gotMode)
	}
	if requiresPassword {
		t.Fatal("expected requiresPassword to be false")
	}
}

func TestDetectMode_WithAuthFile(t *testing.T) {
	// auth/contractor.json があれば Vendor かつパスワード要求になることを確認する。
	dir := t.TempDir()
	authPath := filepath.Join(dir, "auth", "contractor.json")
	if err := os.MkdirAll(filepath.Dir(authPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(authPath, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write auth: %v", err)
	}

	service := NewService(filepath.Join(dir, "ratta.exe"), nil)
	gotMode, requiresPassword, err := service.DetectMode()
	if err != nil {
		t.Fatalf("DetectMode error: %v", err)
	}
	if gotMode != mode.ModeVendor {
		t.Fatalf("unexpected mode: %s", gotMode)
	}
	if !requiresPassword {
		t.Fatal("expected requiresPassword to be true")
	}
}

func TestVerifyContractorPassword_Success(t *testing.T) {
	// 正しいパスワードで Contractor に切り替わることを確認する。
	dir := t.TempDir()
	authPath := filepath.Join(dir, "auth", "contractor.json")
	if err := os.MkdirAll(filepath.Dir(authPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	restore := crypto.SetRandReader(bytes.NewReader(bytes.Repeat([]byte{0x01}, 32)))
	t.Cleanup(restore)

	auth, err := crypto.GenerateContractorAuth("secret")
	if err != nil {
		t.Fatalf("GenerateContractorAuth error: %v", err)
	}
	data, err := jsonfmt.MarshalContractor(auth)
	if err != nil {
		t.Fatalf("MarshalContractor error: %v", err)
	}
	if err := os.WriteFile(authPath, data, 0o644); err != nil {
		t.Fatalf("write auth: %v", err)
	}

	validator, err := schema.NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}
	service := NewService(filepath.Join(dir, "ratta.exe"), validator)
	gotMode, err := service.VerifyContractorPassword("secret")
	if err != nil {
		t.Fatalf("VerifyContractorPassword error: %v", err)
	}
	if gotMode != mode.ModeContractor {
		t.Fatalf("unexpected mode: %s", gotMode)
	}
}

func TestVerifyContractorPassword_WrongPassword(t *testing.T) {
	// 誤ったパスワードでは Contractor にならないことを確認する。
	dir := t.TempDir()
	authPath := filepath.Join(dir, "auth", "contractor.json")
	if err := os.MkdirAll(filepath.Dir(authPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	restore := crypto.SetRandReader(bytes.NewReader(bytes.Repeat([]byte{0x02}, 32)))
	t.Cleanup(restore)

	auth, err := crypto.GenerateContractorAuth("secret")
	if err != nil {
		t.Fatalf("GenerateContractorAuth error: %v", err)
	}
	data, err := jsonfmt.MarshalContractor(auth)
	if err != nil {
		t.Fatalf("MarshalContractor error: %v", err)
	}
	if err := os.WriteFile(authPath, data, 0o644); err != nil {
		t.Fatalf("write auth: %v", err)
	}

	service := NewService(filepath.Join(dir, "ratta.exe"), nil)
	if _, err := service.VerifyContractorPassword("wrong"); err == nil {
		t.Fatal("expected verification error")
	}
}

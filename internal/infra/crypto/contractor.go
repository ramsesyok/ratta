package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	formatVersion    = 1
	kdfName          = "pbkdf2-hmac-sha256"
	kdfIterations    = 200000
	saltSizeBytes    = 16
	nonceSizeBytes   = 16
	derivedKeyLength = 32
)

const fixedPlaintext = "contractor-mode"

// randReader は DD-CLI-005 のランダム生成をテストで固定するための差し替え点。
var randReader io.Reader = rand.Reader

// ContractorAuth は DD-CLI-005 の contractor.json フォーマットを表す。
type ContractorAuth struct {
	FormatVersion int    `json:"format_version"`
	KDF           string `json:"kdf"`
	KDFIterations int    `json:"kdf_iterations"`
	SaltB64       string `json:"salt_b64"`
	NonceB64      string `json:"nonce_b64"`
	CiphertextB64 string `json:"ciphertext_b64"`
	Mode          string `json:"mode"`
}

// GenerateContractorAuth は DD-CLI-005 の方式で contractor.json を生成する。
func GenerateContractorAuth(password string) (ContractorAuth, error) {
	if password == "" {
		return ContractorAuth{}, errors.New("password is required")
	}

	salt := make([]byte, saltSizeBytes)
	if _, err := io.ReadFull(randReader, salt); err != nil {
		return ContractorAuth{}, fmt.Errorf("salt read: %w", err)
	}

	nonce := make([]byte, nonceSizeBytes)
	if _, err := io.ReadFull(randReader, nonce); err != nil {
		return ContractorAuth{}, fmt.Errorf("nonce read: %w", err)
	}

	key := deriveKey(password, salt)
	ciphertext, err := encryptFixed(key, nonce)
	if err != nil {
		return ContractorAuth{}, err
	}

	return ContractorAuth{
		FormatVersion: formatVersion,
		KDF:           kdfName,
		KDFIterations: kdfIterations,
		SaltB64:       base64.StdEncoding.EncodeToString(salt),
		NonceB64:      base64.StdEncoding.EncodeToString(nonce),
		CiphertextB64: base64.StdEncoding.EncodeToString(ciphertext),
		Mode:          "contractor",
	}, nil
}

// VerifyPassword は DD-CLI-005 の固定文字列復号でパスワードを検証する。
func VerifyPassword(auth ContractorAuth, password string) (bool, error) {
	if auth.KDF != kdfName || auth.KDFIterations != kdfIterations {
		return false, fmt.Errorf("unsupported kdf settings")
	}

	salt, err := base64.StdEncoding.DecodeString(auth.SaltB64)
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}
	nonce, err := base64.StdEncoding.DecodeString(auth.NonceB64)
	if err != nil {
		return false, fmt.Errorf("decode nonce: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(auth.CiphertextB64)
	if err != nil {
		return false, fmt.Errorf("decode ciphertext: %w", err)
	}

	key := deriveKey(password, salt)
	plaintext, err := decryptFixed(key, nonce, ciphertext)
	if err != nil {
		return false, nil
	}

	return string(plaintext) == fixedPlaintext, nil
}

// deriveKey は DD-CLI-005 の PBKDF2-HMAC-SHA256 で鍵を導出する。
func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, kdfIterations, derivedKeyLength, sha256.New)
}

// encryptFixed は DD-CLI-005 の固定平文を AES-256-GCM で暗号化する。
func encryptFixed(key, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, nonceSizeBytes)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	return gcm.Seal(nil, nonce, []byte(fixedPlaintext), nil), nil
}

// decryptFixed は DD-CLI-005 の固定平文を AES-256-GCM で復号する。
func decryptFixed(key, nonce, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, nonceSizeBytes)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// SetRandReader は DD-CLI-005 の乱数生成をテスト用に差し替える。
func SetRandReader(reader io.Reader) func() {
	previous := randReader
	randReader = reader
	return func() { randReader = previous }
}

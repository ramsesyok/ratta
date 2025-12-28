package id

import (
	"errors"
	"regexp"
	"testing"

	"github.com/google/uuid"
)

func withDeterministicNanoGenerator(t *testing.T) {
	t.Helper()
	previous := nanoidGenerate
	counter := 0
	nanoidGenerate = func(alphabet string, size int) (string, error) {
		if size <= 0 {
			return "", errors.New("size must be positive")
		}
		base := len(alphabet)
		value := counter
		counter++
		output := make([]byte, size)
		for i := size - 1; i >= 0; i-- {
			output[i] = alphabet[value%base]
			value /= base
		}
		return string(output), nil
	}
	t.Cleanup(func() { nanoidGenerate = previous })
}

func TestNanoIDs_FormatAndUniqueness(t *testing.T) {
	withDeterministicNanoGenerator(t)

	pattern := regexp.MustCompile(`^[A-Za-z0-9_-]{9}$`)
	seen := make(map[string]struct{})

	for i := 0; i < 100; i++ {
		value, err := NewIssueID()
		if err != nil {
			t.Fatalf("NewIssueID error: %v", err)
		}
		if !pattern.MatchString(value) {
			t.Fatalf("unexpected issue id format: %s", value)
		}
		if _, exists := seen[value]; exists {
			t.Fatalf("issue id is not unique: %s", value)
		}
		seen[value] = struct{}{}
	}

	attachmentID, err := NewAttachmentID()
	if err != nil {
		t.Fatalf("NewAttachmentID error: %v", err)
	}
	if !pattern.MatchString(attachmentID) {
		t.Fatalf("unexpected attachment id format: %s", attachmentID)
	}
}

func TestCommentID_FormatAndUniqueness(t *testing.T) {
	previous := uuidV7Generator
	defer func() { uuidV7Generator = previous }()

	counter := byte(1)
	uuidV7Generator = func() (uuid.UUID, error) {
		var data [16]byte
		data[15] = counter
		counter++
		data[6] = (data[6] & 0x0f) | 0x70
		data[8] = (data[8] & 0x3f) | 0x80
		return uuid.FromBytes(data[:])
	}

	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	first, err := NewCommentID()
	if err != nil {
		t.Fatalf("NewCommentID error: %v", err)
	}
	second, err := NewCommentID()
	if err != nil {
		t.Fatalf("NewCommentID error: %v", err)
	}
	if first == second {
		t.Fatalf("comment id is not unique: %s", first)
	}
	if !pattern.MatchString(first) {
		t.Fatalf("unexpected comment id format: %s", first)
	}
	if !pattern.MatchString(second) {
		t.Fatalf("unexpected comment id format: %s", second)
	}
}

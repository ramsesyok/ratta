package jsonfmt

import "testing"

func TestMarshalCanonicalIndentation(t *testing.T) {
	// JSON が 2 スペースのインデントと LF 改行で出力されることを確認する。
	type sample struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	data, err := MarshalCanonical(sample{Name: "alpha", Count: 2})
	if err != nil {
		t.Fatalf("MarshalCanonical error: %v", err)
	}

	expected := "{\n  \"count\": 2,\n  \"name\": \"alpha\"\n}\n"
	if string(data) != expected {
		t.Fatalf("unexpected JSON output:\n%s", string(data))
	}
}

func TestMarshalIssue_KeyOrder(t *testing.T) {
	// issue JSON のキー順が DD-DATA-003/004/005 に沿っていることを確認する。
	input := map[string]any{
		"status":         "Open",
		"issue_id":       "ABC123def",
		"version":        1,
		"category":       "alpha",
		"title":          "Title",
		"description":    "Desc",
		"priority":       "High",
		"origin_company": "Vendor",
		"assignee":       "User",
		"created_at":     "2024-01-01T00:00:00Z",
		"updated_at":     "2024-01-02T00:00:00Z",
		"due_date":       "2024-01-03",
		"comments": []any{
			map[string]any{
				"author_company": "Vendor",
				"comment_id":     "00000000-0000-7000-8000-000000000001",
				"body":           "Note",
				"author_name":    "User",
				"created_at":     "2024-01-02T00:00:00Z",
				"attachments": []any{
					map[string]any{
						"stored_name":   "ATTACH_x.txt",
						"attachment_id": "ATTACH123",
						"relative_path": "ABC123def.files/ATTACH_x.txt",
						"file_name":     "x.txt",
						"mime_type":     "text/plain",
						"size_bytes":    12,
					},
				},
			},
		},
	}

	got, err := MarshalIssue(input)
	if err != nil {
		t.Fatalf("MarshalIssue error: %v", err)
	}

	expected := "{\n" +
		"  \"version\": 1,\n" +
		"  \"issue_id\": \"ABC123def\",\n" +
		"  \"category\": \"alpha\",\n" +
		"  \"title\": \"Title\",\n" +
		"  \"description\": \"Desc\",\n" +
		"  \"status\": \"Open\",\n" +
		"  \"priority\": \"High\",\n" +
		"  \"origin_company\": \"Vendor\",\n" +
		"  \"assignee\": \"User\",\n" +
		"  \"created_at\": \"2024-01-01T00:00:00Z\",\n" +
		"  \"updated_at\": \"2024-01-02T00:00:00Z\",\n" +
		"  \"due_date\": \"2024-01-03\",\n" +
		"  \"comments\": [\n" +
		"    {\n" +
		"      \"comment_id\": \"00000000-0000-7000-8000-000000000001\",\n" +
		"      \"body\": \"Note\",\n" +
		"      \"author_name\": \"User\",\n" +
		"      \"author_company\": \"Vendor\",\n" +
		"      \"created_at\": \"2024-01-02T00:00:00Z\",\n" +
		"      \"attachments\": [\n" +
		"        {\n" +
		"          \"attachment_id\": \"ATTACH123\",\n" +
		"          \"file_name\": \"x.txt\",\n" +
		"          \"stored_name\": \"ATTACH_x.txt\",\n" +
		"          \"relative_path\": \"ABC123def.files/ATTACH_x.txt\",\n" +
		"          \"mime_type\": \"text/plain\",\n" +
		"          \"size_bytes\": 12\n" +
		"        }\n" +
		"      ]\n" +
		"    }\n" +
		"  ]\n" +
		"}\n"

	if string(got) != expected {
		t.Fatalf("unexpected issue JSON:\n%s", string(got))
	}
}

func TestMarshalConfig_KeyOrder(t *testing.T) {
	// config JSON のキー順が DD-DATA-001 に沿っていることを確認する。
	input := map[string]any{
		"ui": map[string]any{
			"page_size": 20,
		},
		"format_version":         1,
		"last_project_root_path": "C:/proj",
		"log": map[string]any{
			"level": "info",
		},
	}

	got, err := MarshalConfig(input)
	if err != nil {
		t.Fatalf("MarshalConfig error: %v", err)
	}

	expected := "{\n" +
		"  \"format_version\": 1,\n" +
		"  \"last_project_root_path\": \"C:/proj\",\n" +
		"  \"log\": {\n" +
		"    \"level\": \"info\"\n" +
		"  },\n" +
		"  \"ui\": {\n" +
		"    \"page_size\": 20\n" +
		"  }\n" +
		"}\n"
	if string(got) != expected {
		t.Fatalf("unexpected config JSON:\n%s", string(got))
	}
}

func TestMarshalContractor_KeyOrder(t *testing.T) {
	// contractor JSON のキー順が DD-DATA-001 に沿っていることを確認する。
	input := map[string]any{
		"mode":           "contractor",
		"ciphertext_b64": "cc",
		"salt_b64":       "aa",
		"nonce_b64":      "bb",
		"kdf":            "pbkdf2-hmac-sha256",
		"kdf_iterations": 200000,
		"format_version": 1,
	}

	got, err := MarshalContractor(input)
	if err != nil {
		t.Fatalf("MarshalContractor error: %v", err)
	}

	expected := "{\n" +
		"  \"format_version\": 1,\n" +
		"  \"kdf\": \"pbkdf2-hmac-sha256\",\n" +
		"  \"kdf_iterations\": 200000,\n" +
		"  \"salt_b64\": \"aa\",\n" +
		"  \"nonce_b64\": \"bb\",\n" +
		"  \"ciphertext_b64\": \"cc\",\n" +
		"  \"mode\": \"contractor\"\n" +
		"}\n"
	if string(got) != expected {
		t.Fatalf("unexpected contractor JSON:\n%s", string(got))
	}
}

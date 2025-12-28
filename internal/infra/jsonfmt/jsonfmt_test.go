package jsonfmt

import "testing"

func TestMarshalCanonicalIndentation(t *testing.T) {
	type sample struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	data, err := MarshalCanonical(sample{Name: "alpha", Count: 2})
	if err != nil {
		t.Fatalf("MarshalCanonical error: %v", err)
	}

	expected := "{\n  \"name\": \"alpha\",\n  \"count\": 2\n}\n"
	if string(data) != expected {
		t.Fatalf("unexpected JSON output:\n%s", string(data))
	}
}

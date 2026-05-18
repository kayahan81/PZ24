package handler

import (
	"testing"
)

func TestSanitizeDescription(t *testing.T) {
	input := "<script>alert(1)</script>"
	expected := "&lt;script&gt;alert(1)&lt;/script&gt;"

	result := sanitizeDescription(input)

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

package control

import (
	"strings"
	"testing"
)

func TestTruncateSlackTextHonorsMax(t *testing.T) {
	input := strings.Repeat("x", 100)

	got := truncateSlackText(input, 64)
	if len(got) != 64 {
		t.Fatalf("expected truncated text length 64, got %d", len(got))
	}
	if !strings.HasSuffix(got, dbReadSlackTruncated) {
		t.Fatalf("expected truncation suffix")
	}
}

func TestTruncateSlackTextShortMaxHonorsMax(t *testing.T) {
	input := strings.Repeat("x", 100)

	got := truncateSlackText(input, 8)
	if len(got) != 8 {
		t.Fatalf("expected truncated text length 8, got %d", len(got))
	}
}

package store

import (
	"math"
	"strings"
	"testing"
	"time"
)

type executionLedgerScannerFunc func(dest ...any) error

func (f executionLedgerScannerFunc) Scan(dest ...any) error {
	return f(dest...)
}

func TestScanExecutionLedgerEventRejectsInvalidPayloadJSON(t *testing.T) {
	scanner := executionLedgerScannerFunc(func(dest ...any) error {
		*(dest[0].(*string)) = "xled-invalid"
		*(dest[1].(*string)) = "hexec-1"
		*(dest[2].(*string)) = "op-1"
		*(dest[3].(*string)) = "trace-1"
		*(dest[4].(*string)) = "workflow-1"
		*(dest[5].(*string)) = "main"
		*(dest[6].(*string)) = "tool.call.completed"
		*(dest[7].(*string)) = "completed"
		*(dest[8].(*int)) = 1
		*(dest[9].(*string)) = "idem-1"
		*(dest[10].(*[]byte)) = []byte(`{"unterminated":`)
		*(dest[11].(*time.Time)) = time.Now().UTC()
		return nil
	})

	_, err := scanExecutionLedgerEvent(scanner)
	if err == nil {
		t.Fatal("expected invalid ledger payload JSON to return an error")
	}
	if !strings.Contains(err.Error(), "decode execution ledger payload for xled-invalid") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJsonStringPanicsOnUnsupportedValue(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("expected unsupported JSON value to panic")
		}
	}()

	_ = jsonString(map[string]any{"score": math.NaN()})
}

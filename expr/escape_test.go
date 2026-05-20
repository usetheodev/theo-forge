package expr

import (
	"strings"
	"testing"
)

// T2.2 regression tests (SEC-002, sec-002-expr-injection-confirmed).
// Covers ADR-005 (argoEscape) and EC-4 (double-escape contract).

func TestArgoEscape_DoublesSingleQuotes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no quote", "plain", "plain"},
		{"single quote", "it's", "it''s"},
		{"multiple quotes", "a'b'c", "a''b''c"},
		{"only quote", "'", "''"},
		{"empty", "", ""},
		{"unicode preserved", "café's", "café''s"},
		{"already double escaped becomes quadruple (contract)", "it''s", "it''''s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := argoEscape(tt.in); got != tt.want {
				t.Errorf("argoEscape(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestExprC_EscapesSingleQuotes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "plain", "'plain'"},
		{"injection attempt", "main' || 'true", "'main'' || ''true'"},
		{"with apostrophe", "it's", "'it''s'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := C(tt.in).String()
			if got != tt.want {
				t.Errorf("C(%q).String() = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// EC-4: C() always escapes; callers passing pre-escaped strings double-escape.
// Migration path is RawC. Test locks the contract.
func TestExprC_DoubleEscapesPreEscapedInput(t *testing.T) {
	got := C("it''s").String()
	want := "'it''''s'"
	if got != want {
		t.Fatalf("C(\"it''s\").String() = %q, want %q (use RawC for pre-escaped inputs)", got, want)
	}
}

func TestExprRawC_NoEscape(t *testing.T) {
	got := RawC("already''escaped").String()
	want := "'already''escaped'"
	if got != want {
		t.Fatalf("RawC().String() = %q, want %q", got, want)
	}
}

func TestExprContains_Escapes(t *testing.T) {
	got := E("name").Contains("it's").String()
	if !strings.Contains(got, "it''s") {
		t.Fatalf("Contains escape lost: got %q", got)
	}
}

func TestExprMatches_Escapes(t *testing.T) {
	got := E("name").Matches("^[a-z']+$").String()
	if !strings.Contains(got, "[a-z'']+") {
		t.Fatalf("Matches escape lost: got %q", got)
	}
}

func TestExprStartsWith_Escapes(t *testing.T) {
	got := E("name").StartsWith("o'").String()
	if !strings.Contains(got, "o''") {
		t.Fatalf("StartsWith escape lost: got %q", got)
	}
}

func TestExprEndsWith_Escapes(t *testing.T) {
	got := E("name").EndsWith("'s").String()
	if !strings.Contains(got, "''s") {
		t.Fatalf("EndsWith escape lost: got %q", got)
	}
}

func TestSprig_Escapes(t *testing.T) {
	if got := Sprig.Trim("it's").String(); !strings.Contains(got, "it''s") {
		t.Errorf("Sprig.Trim escape lost: %q", got)
	}
	if got := Sprig.Upper("it's").String(); !strings.Contains(got, "it''s") {
		t.Errorf("Sprig.Upper escape lost: %q", got)
	}
	if got := Sprig.Lower("it's").String(); !strings.Contains(got, "it''s") {
		t.Errorf("Sprig.Lower escape lost: %q", got)
	}
	got := Sprig.Replace("o'", "x", "it's").String()
	if !strings.Contains(got, "o''") || !strings.Contains(got, "it''s") {
		t.Errorf("Sprig.Replace escape lost: %q", got)
	}
}

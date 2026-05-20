package expr

import (
	"strings"
	"testing"
)

// FuzzExprC asserts the structural invariants of expr.C against arbitrary
// inputs. (Extra-1 — historical CVE site SEC-002.)
//
// Invariants:
//  1. The repr always begins with `'` and ends with `'`.
//  2. The repr NEVER contains an unescaped single quote in its interior.
//     "Unescaped" means: a `'` that is not part of an adjacent `”`
//     pair AND that is not the opening/closing wrapper.
//  3. argoEscape is idempotent on its OWN output: escaping an
//     already-escaped string twice produces 4x the doubled quotes,
//     never breaks parity.
//
// Seed corpus includes the regression inputs from the SEC-002 review.
func FuzzExprC(f *testing.F) {
	for _, seed := range []string{
		"",
		"plain",
		"it's",
		"a'b'c",
		"main' || 'true",          // the SEC-002 PoC: injection attempt
		"''",                      // pre-escaped (legitimate footgun)
		`"double"`,                // mixed quotes
		"emoji 🐹",                 // multi-byte
		"\x00null",                // NUL byte
		strings.Repeat("a'b", 50), // many escapes
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, in string) {
		e := C(in)
		repr := e.String()

		// (1) Wrapping invariant. All non-string inputs go through other
		// branches of C(); the fuzz signature constrains us to strings so
		// the wrap must always be present.
		if !strings.HasPrefix(repr, "'") || !strings.HasSuffix(repr, "'") {
			t.Fatalf("repr lost wrapping for input %q: %q", in, repr)
		}

		// (2) Interior must have every `'` doubled. Strip the outer wrappers
		// and count: every standalone `'` must be part of an adjacent pair.
		interior := repr[1 : len(repr)-1]
		for i := 0; i < len(interior); i++ {
			if interior[i] != '\'' {
				continue
			}
			// `interior[i]` is `'`. The next char MUST also be `'`.
			if i+1 >= len(interior) || interior[i+1] != '\'' {
				t.Fatalf("found unescaped quote at offset %d in %q (input was %q)",
					i, repr, in)
			}
			i++ // skip the partner
		}

		// (3) argoEscape idempotency-on-pair-count. For every `'` in
		// the input, the interior must contain exactly TWO `'` (the doubled form).
		wantQuotes := 2 * strings.Count(in, "'")
		gotQuotes := strings.Count(interior, "'")
		if wantQuotes != gotQuotes {
			t.Fatalf("quote count mismatch for input %q: want %d, got %d (interior=%q)",
				in, wantQuotes, gotQuotes, interior)
		}
	})
}

// FuzzExprContains covers the same surface for the string method helpers
// that ALSO interpolate via single quotes (Contains, Matches, StartsWith,
// EndsWith, Sprig.*). One representative is enough — they share argoEscape.
func FuzzExprContains(f *testing.F) {
	for _, seed := range []string{"", "x", "it's", "''"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, needle string) {
		base := E("x")
		repr := base.Contains(needle).String()
		// Format MUST match exactly: `x.contains('<escaped>')`.
		const prefix = "x.contains('"
		const suffix = "')"
		if !strings.HasPrefix(repr, prefix) || !strings.HasSuffix(repr, suffix) {
			t.Fatalf("unexpected shape for needle %q: %q", needle, repr)
		}
		interior := repr[len(prefix) : len(repr)-len(suffix)]
		wantQuotes := 2 * strings.Count(needle, "'")
		if got := strings.Count(interior, "'"); got != wantQuotes {
			t.Fatalf("Contains escape lost parity for %q: want %d, got %d",
				needle, wantQuotes, got)
		}
	})
}

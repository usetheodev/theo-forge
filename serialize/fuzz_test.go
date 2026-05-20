package serialize

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// FuzzWorkflowToFile_PathTraversal asserts that no input to the
// `fileName` argument of WorkflowToFile can produce a write outside
// the supplied output directory. (Extra-1 — historical CVE site SEC-001.)
//
// Strategy: fuzz fileName, run WorkflowToFile against a fresh TempDir,
// then verify any returned path is strictly under TempDir. A failure
// either:
//
//	a) returns a path NOT under TempDir → containment broken (CVE);
//	b) writes to a path NOT under TempDir → same, even worse;
//	c) panics → also a bug (input validation crash).
//
// We accept errors as the correct response for any pathological input.
func FuzzWorkflowToFile_PathTraversal(f *testing.F) {
	for _, seed := range []string{
		"file.yaml",
		"../escape",
		"../../etc/passwd",
		"/absolute",
		"",
		".",
		"..",
		"sub/../../escape",
		"sub/legit.yaml",
		string([]byte{0}),        // NUL byte
		strings.Repeat("a", 300), // long
		"name\nwith\nnewlines",   // control chars
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, fileName string) {
		dir := t.TempDir()
		absDir, err := filepath.Abs(dir)
		if err != nil {
			t.Skip("cannot resolve TempDir to abs")
		}

		path, err := WorkflowToFile("apiVersion: argoproj.io/v1alpha1\nkind: Workflow\n",
			dir, fileName, "name-from-arg", "")
		if err != nil {
			// Rejection is the correct outcome for malformed input.
			// Confirm it is an ErrPathTraversal (sentinel) — opaque
			// errors are also acceptable, but the sentinel proves the
			// containment helper fired and not some other failure.
			if errors.Is(err, model.ErrPathTraversal) {
				return
			}
			// Any other error (e.g., empty fileName + empty derived name)
			// is also acceptable — the goal is "no path escape".
			return
		}

		// Success path: the returned path MUST be inside absDir.
		absPath, err := filepath.Abs(path)
		if err != nil {
			t.Fatalf("returned path %q not absolute-resolvable: %v", path, err)
		}
		// Resolve symlinks before comparing to defeat a future regression
		// where the parent dir is a symlink (EC-3 territory).
		realDir, derr := filepath.EvalSymlinks(absDir)
		if derr == nil {
			absDir = realDir
		}
		realPath, perr := filepath.EvalSymlinks(absPath)
		if perr == nil {
			absPath = realPath
		}
		rel, err := filepath.Rel(absDir, absPath)
		if err != nil {
			t.Fatalf("Rel(%q, %q) failed: %v", absDir, absPath, err)
		}
		// rel CAN start with the bytes "..": a file literally named
		// "..yaml" is legitimate. We only flag traversal when rel ==
		// ".." (the parent itself) OR starts with "../" / "..\\"
		// (a path component escaping outward).
		sep := string(filepath.Separator)
		if rel == ".." || strings.HasPrefix(rel, ".."+sep) {
			t.Fatalf("path traversal escaped: fileName=%q -> wrote to %q (relative to %q: %q)",
				fileName, absPath, absDir, rel)
		}

		// Sanity: the file actually exists where the helper says.
		if _, err := os.Stat(absPath); err != nil {
			t.Fatalf("returned path %q does not exist on disk: %v", absPath, err)
		}
	})
}

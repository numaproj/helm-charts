package mirror

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTokenizeDetokenizeRoundtrip_AllMirroredFiles(t *testing.T) {
	// Round-trip every real mirrored chart file currently in the repo.
	// This is the hard invariant: Detokenize(Tokenize(x)) must equal x.
	chartRoot, err := chartRootForTest()
	if err != nil {
		t.Fatalf("locating chart root: %v", err)
	}
	for _, m := range MirroredFiles {
		path := filepath.Join(chartRoot, m.LocalPath)
		original, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("read %s: %v", m.LocalPath, err)
			continue
		}
		tokenized, subs := Tokenize(original)
		restored := Detokenize(tokenized, subs)
		if !bytes.Equal(original, restored) {
			t.Errorf("roundtrip mismatch for %s\n--- original ---\n%s\n--- restored ---\n%s",
				m.LocalPath, hexish(original), hexish(restored))
		}
	}
}

func TestTokenize_ProducesNoTemplateRuns(t *testing.T) {
	// After tokenisation no `{{` or `}}` should remain — that would mean a
	// template construct slipped past the regex and would confuse git's
	// line-based merge.
	chartRoot, err := chartRootForTest()
	if err != nil {
		t.Fatalf("locating chart root: %v", err)
	}
	for _, m := range MirroredFiles {
		original, err := os.ReadFile(filepath.Join(chartRoot, m.LocalPath))
		if err != nil {
			t.Errorf("read %s: %v", m.LocalPath, err)
			continue
		}
		tokenized, _ := Tokenize(original)
		if bytes.Contains(tokenized, []byte("{{")) || bytes.Contains(tokenized, []byte("}}")) {
			t.Errorf("residual template markers in tokenized %s: %s", m.LocalPath, tokenized)
		}
	}
}

func TestTokenize_Inline(t *testing.T) {
	src := []byte(`key: {{ .Values.X | quote }}`)
	tok, subs := Tokenize(src)
	if bytes.Contains(tok, []byte("{{")) {
		t.Fatalf("inline expression survived tokenisation: %s", tok)
	}
	if !bytes.HasPrefix(tok, []byte("key: __NFTPLEXPR_")) {
		t.Fatalf("inline sentinel form unexpected: %s", tok)
	}
	if !bytes.Equal(Detokenize(tok, subs), src) {
		t.Fatalf("roundtrip mismatch for inline")
	}
}

func TestTokenize_BlockDirective(t *testing.T) {
	src := []byte(`  {{- if .Values.dexServer.enabled }}` + "\n" +
		`  key: value` + "\n" +
		`  {{- end }}`)
	tok, subs := Tokenize(src)
	if bytes.Contains(tok, []byte("{{")) {
		t.Fatalf("block expression survived tokenisation: %s", tok)
	}
	if !bytes.Contains(tok, []byte("  # __NFTPLLINE_")) {
		t.Fatalf("expected block sentinel as YAML comment with indent: %s", tok)
	}
	if !bytes.Equal(Detokenize(tok, subs), src) {
		t.Fatalf("roundtrip mismatch for block")
	}
}

func TestTokenize_MultipleInlinePerLine(t *testing.T) {
	src := []byte(`server.address: "{{ .Values.server.configs.host }}:{{ include "server.configs.port" . }}"`)
	tok, subs := Tokenize(src)
	count := bytes.Count(tok, []byte("__NFTPLEXPR_"))
	if count != 2 {
		t.Fatalf("want 2 inline sentinels, got %d in %s", count, tok)
	}
	if !bytes.Equal(Detokenize(tok, subs), src) {
		t.Fatalf("roundtrip mismatch for multi-inline")
	}
}

func TestTokenize_LeavesPlainTextAlone(t *testing.T) {
	src := []byte(`apiVersion: v1` + "\n" + `kind: ConfigMap` + "\n" + `# this is a comment`)
	tok, _ := Tokenize(src)
	if !bytes.Equal(tok, src) {
		t.Fatalf("plain text was modified:\n--- before ---\n%s\n--- after ---\n%s", src, tok)
	}
}

// chartRootForTest finds charts/numaflow/ by walking up from CWD.
// Tests run with CWD = upgrade/internal/mirror, so we walk up to repo root.
func chartRootForTest() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		candidate := filepath.Join(dir, "charts", "numaflow")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// hexish renders a byte slice with visible markers for whitespace so a
// roundtrip mismatch is easy to read in test output.
func hexish(b []byte) string {
	s := string(b)
	s = strings.ReplaceAll(s, "\t", "→")
	s = strings.ReplaceAll(s, " ", "·")
	return s
}

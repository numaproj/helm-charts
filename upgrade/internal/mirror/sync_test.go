package mirror

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

// TestRun_GuardsFromEqualsTo pins the fix for the "default --from-version
// collides with --to-version" blocker: when the two resolve to the same
// value the run must fail loudly instead of reporting every file no-change.
// The guard fires before any network call, so this test needs no upstream
// access. It also confirms both versions are normalised to a leading `v`.
func TestRun_GuardsFromEqualsTo(t *testing.T) {
	var buf bytes.Buffer
	_, err := Run(SyncOptions{FromVersion: "v9.9.9", ToVersion: "9.9.9", Out: &buf})
	if err == nil {
		t.Fatalf("expected from==to guard to fail the run, got nil error")
	}
	if !strings.Contains(err.Error(), "equals") {
		t.Fatalf("unexpected guard error: %v", err)
	}
}

// TestAppVersionFromGitHEAD exercises the HEAD-based default resolution
// against the real repo. It should return the appVersion committed to HEAD's
// Chart.yaml (non-empty, no `v` prefix stripping needed here — that happens in
// Run). Skips when git or the committed chart file is unavailable.
func TestAppVersionFromGitHEAD(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not on PATH: %v", err)
	}
	chartRoot, err := chartRootForTest()
	if err != nil {
		t.Fatalf("locating chart root: %v", err)
	}
	got, err := appVersionFromGitHEAD(chartRoot)
	if err != nil {
		t.Skipf("HEAD Chart.yaml not readable (shallow clone / detached tree?): %v", err)
	}
	if strings.TrimSpace(got) == "" {
		t.Fatalf("appVersionFromGitHEAD returned empty appVersion")
	}
}

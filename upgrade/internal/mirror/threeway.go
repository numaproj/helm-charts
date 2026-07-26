package mirror

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MergeOutcome describes the result of merging upstream changes onto the
// chart's templated copy.
type MergeOutcome int

const (
	MergeOutcomeUnknown MergeOutcome = iota
	// MergeClean means the upstream change applied without conflict.
	MergeClean
	// MergeConflict means the merge produced conflict markers; the maintainer
	// must resolve them by hand before applying the result to the chart.
	MergeConflict
	// MergeErrored means the underlying tooling failed for an unexpected
	// reason (binary missing, IO error, etc).
	MergeErrored
)

func (o MergeOutcome) String() string {
	switch o {
	case MergeClean:
		return "clean-merge"
	case MergeConflict:
		return "conflict"
	case MergeErrored:
		return "error"
	default:
		return "unknown"
	}
}

// MergeResult is the L2 output for a single mirrored file.
type MergeResult struct {
	Outcome      MergeOutcome
	Merged       []byte // detokenized result; contains conflict markers when Outcome == MergeConflict
	NumConflicts int    // number of conflict regions reported by git
	Err          error
}

// ThreeWayMerge applies the upstream change (base -> theirs) onto the chart's
// templated copy. The chart copy is tokenized first so Helm template
// constructs do not match against upstream literals.
//
// To avoid false-positive conflicts where chart-only additions (labels,
// namespace, includes) sit adjacent to upstream changes, we first compute the
// chart's "pure addition" hunks against base — content the chart adds that
// upstream has no opinion on — and apply those hunks to both base and theirs.
// The 3-way merge then sees inputs where the chart-only structure exists on
// all three sides, so genuine adjacency is not mistaken for an edit conflict.
//
// Real conflicts (chart values a line that upstream is also changing) still
// produce `<<<<<<< / ||||||| / ======= / >>>>>>>` markers in the output.
//
// Call EnsureToolsAvailable once before iterating the registry so missing
// binaries surface up front.
func ThreeWayMerge(chartCopy, base, theirs []byte) MergeResult {
	if bytes.Equal(base, theirs) {
		tok, subs := Tokenize(chartCopy)
		return MergeResult{Outcome: MergeClean, Merged: Detokenize(tok, subs)}
	}

	tokenized, subs := Tokenize(chartCopy)

	scratch, err := os.MkdirTemp("", "mirror-merge-*")
	if err != nil {
		return MergeResult{Outcome: MergeErrored, Err: fmt.Errorf("create scratch dir: %w", err)}
	}
	defer os.RemoveAll(scratch)

	// Step 1: extract the chart's pure-addition hunks against base. These are
	// the chart-only lines (labels, namespace, includes, conditional blocks)
	// that have no upstream counterpart.
	addPatch, err := chartOnlyAdditionsPatch(scratch, base, tokenized)
	if err != nil {
		return MergeResult{Outcome: MergeErrored, Err: err}
	}

	// Step 2: enrich base and theirs by applying the chart-only additions to
	// both, producing base' and theirs' that share the chart's structural
	// additions. If the patch is empty (no chart-only additions), the inputs
	// pass through unchanged.
	basePrime, err := applyAdditionsPatch(scratch, "base", base, addPatch)
	if err != nil {
		return MergeResult{Outcome: MergeErrored, Err: err}
	}
	theirsPrime, err := applyAdditionsPatch(scratch, "theirs", theirs, addPatch)
	if err != nil {
		return MergeResult{Outcome: MergeErrored, Err: err}
	}

	// Step 3: run git merge-file --diff3. Any conflict surfaced here reflects
	// a real disagreement (chart and upstream both editing the same line).
	merged, conflicts, runErr := runGitMergeFile(scratch, tokenized, basePrime, theirsPrime)
	if runErr != nil {
		return MergeResult{Outcome: MergeErrored, Err: runErr}
	}

	merged = Detokenize(merged, subs)
	if conflicts > 0 {
		return MergeResult{Outcome: MergeConflict, Merged: merged, NumConflicts: conflicts}
	}
	return MergeResult{Outcome: MergeClean, Merged: merged}
}

// chartOnlyAdditionsPatch computes the diff base -> tokenized and keeps only
// hunks consisting purely of additions (no `-` lines). The returned bytes are
// a valid unified diff that can be fed to `patch` to replay the chart's
// structural additions onto another file.
func chartOnlyAdditionsPatch(scratch string, base, tokenized []byte) ([]byte, error) {
	basePath := filepath.Join(scratch, "base.in")
	oursPath := filepath.Join(scratch, "ours.in")
	if err := os.WriteFile(basePath, base, 0o600); err != nil {
		return nil, fmt.Errorf("write base.in: %w", err)
	}
	if err := os.WriteFile(oursPath, tokenized, 0o600); err != nil {
		return nil, fmt.Errorf("write ours.in: %w", err)
	}

	// Use -U 0 so each contiguous edit region becomes its own hunk. That way
	// a hunk containing only `+` lines is a true pure-addition cluster, even
	// when it sits adjacent to a modification hunk in the chart. With the
	// default context width, an addition next to a modification gets merged
	// into a single hunk and the filter would drop both halves.
	cmd := exec.Command("diff", "-U", "0", basePath, oursPath)
	out, err := cmd.Output()
	// `diff` exits 1 when files differ (the normal case) and 0 when
	// identical. We only care about exit codes >= 2 (real error).
	if err != nil {
		if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() >= 2 {
			return nil, fmt.Errorf("diff failed: %w", err)
		}
	}
	if len(out) == 0 {
		return nil, nil
	}
	return filterAdditionsOnly(out), nil
}

// filterAdditionsOnly walks a unified diff and keeps only hunks that contain
// no `-` (deletion) lines. The file-pair header is preserved so the output is
// still a valid patch.
func filterAdditionsOnly(diff []byte) []byte {
	lines := strings.Split(string(diff), "\n")
	var out []string
	var header []string
	var hunk []string
	var hunkHasDeletion bool

	flushHunk := func() {
		if len(hunk) == 0 {
			return
		}
		if !hunkHasDeletion {
			out = append(out, hunk...)
		}
		hunk = nil
		hunkHasDeletion = false
	}

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ "):
			header = append(header, line)
		case strings.HasPrefix(line, "@@"):
			flushHunk()
			hunk = []string{line}
		case len(hunk) > 0:
			if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "--- ") {
				hunkHasDeletion = true
			}
			hunk = append(hunk, line)
		}
	}
	flushHunk()

	if len(out) == 0 {
		return nil
	}
	return []byte(strings.Join(append(header, out...), "\n"))
}

// applyAdditionsPatch applies the chart's additions to one of base/theirs,
// producing the enriched input for git merge-file. An empty patch passes the
// input through unchanged.
//
// Rejected hunks (`patch` exit 1) are treated as best-effort enrichment —
// some additions did not anchor — and we continue with whatever patch
// produced rather than failing the whole merge. A genuinely broken patch
// run (exit >= 2) is escalated.
func applyAdditionsPatch(scratch, label string, target, patch []byte) ([]byte, error) {
	if len(patch) == 0 {
		return target, nil
	}
	targetPath := filepath.Join(scratch, label+".enriched")
	patchPath := filepath.Join(scratch, label+".addpatch")
	rejPath := filepath.Join(scratch, label+".addrej")
	if err := os.WriteFile(targetPath, target, 0o600); err != nil {
		return nil, fmt.Errorf("write %s: %w", targetPath, err)
	}
	if err := os.WriteFile(patchPath, patch, 0o600); err != nil {
		return nil, fmt.Errorf("write %s: %w", patchPath, err)
	}
	cmd := exec.Command("patch", "--quiet", "-r", rejPath, "-i", patchPath, targetPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if runErr != nil {
		if exit, ok := runErr.(*exec.ExitError); !ok || exit.ExitCode() >= 2 {
			return nil, fmt.Errorf("patch %s: %w (stderr: %s)", label, runErr, stderr.String())
		}
		// Exit 1: some hunks rejected. Continue with best-effort enrichment.
	}
	return os.ReadFile(targetPath)
}

// runGitMergeFile executes git merge-file --diff3 with the supplied inputs,
// returning the merged bytes and the number of conflict regions.
func runGitMergeFile(scratch string, ours, base, theirs []byte) ([]byte, int, error) {
	oursPath := filepath.Join(scratch, "merge.ours")
	basePath := filepath.Join(scratch, "merge.base")
	theirsPath := filepath.Join(scratch, "merge.theirs")
	for _, p := range []struct {
		path string
		data []byte
	}{{oursPath, ours}, {basePath, base}, {theirsPath, theirs}} {
		if err := os.WriteFile(p.path, p.data, 0o600); err != nil {
			return nil, 0, fmt.Errorf("write %s: %w", p.path, err)
		}
	}
	cmd := exec.Command("git", "merge-file", "--diff3",
		"-L", "ours", "-L", "base", "-L", "theirs",
		oursPath, basePath, theirsPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	merged, readErr := os.ReadFile(oursPath)
	if readErr != nil {
		return nil, 0, fmt.Errorf("read merge output: %w", readErr)
	}

	if runErr != nil {
		exit, ok := runErr.(*exec.ExitError)
		if !ok {
			return nil, 0, fmt.Errorf("git merge-file: %w (stderr: %s)", runErr, stderr.String())
		}
		code := exit.ExitCode()
		if code < 0 {
			return nil, 0, fmt.Errorf("git merge-file failed: %s", stderr.String())
		}
		return merged, code, nil
	}
	return merged, 0, nil
}

// EnsureToolsAvailable verifies that `git`, `diff`, and `patch` are on PATH.
// Callers should invoke this once before iterating the registry so failures
// surface up front rather than partway through a sync run.
func EnsureToolsAvailable() error {
	for _, tool := range []string{"git", "diff", "patch"} {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s is required for `sync`: %w", tool, err)
		}
	}
	return nil
}

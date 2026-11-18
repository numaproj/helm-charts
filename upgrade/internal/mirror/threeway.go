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
	// that have no upstream counterpart. Additions that already exist in
	// theirs are upstream changes the chart has absorbed; replaying them
	// onto theirs corrupts the merge, so they are left out.
	addPatch, err := chartOnlyAdditionsPatch(scratch, base, tokenized, theirs)
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
// hunks consisting purely of additions (no `-` lines). Additions whose lines
// already appear contiguously in theirs, and not in base, are upstream
// changes the chart has already taken; they are omitted so a later re-check
// does not replay them. The returned bytes are a valid unified diff that can
// be fed to `patch`.
func chartOnlyAdditionsPatch(scratch string, base, tokenized, theirs []byte) ([]byte, error) {
	out, err := unifiedDiff(scratch, "base.in", "ours.in", base, tokenized)
	if err != nil || len(out) == 0 {
		return nil, err
	}
	return filterChartOnlyAdditions(out, base, theirs), nil
}

// unifiedDiff runs `diff -U 0`. Exit code 1 (files differ) is success.
func unifiedDiff(scratch, nameA, nameB string, a, b []byte) ([]byte, error) {
	pathA := filepath.Join(scratch, nameA)
	pathB := filepath.Join(scratch, nameB)
	if err := os.WriteFile(pathA, a, 0o600); err != nil {
		return nil, fmt.Errorf("write %s: %w", nameA, err)
	}
	if err := os.WriteFile(pathB, b, 0o600); err != nil {
		return nil, fmt.Errorf("write %s: %w", nameB, err)
	}

	// Use -U 0 so each contiguous edit region becomes its own hunk. That way
	// a hunk containing only `+` lines is a true pure-addition cluster, even
	// when it sits adjacent to a modification hunk in the chart. With the
	// default context width, an addition next to a modification gets merged
	// into a single hunk and the filter would drop both halves.
	cmd := exec.Command("diff", "-U", "0", pathA, pathB)
	out, err := cmd.Output()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() >= 2 {
			return nil, fmt.Errorf("diff failed: %w", err)
		}
	}
	return out, nil
}

// filterAdditionsOnly walks a unified diff and keeps only hunks that contain
// no `-` (deletion) lines.
func filterAdditionsOnly(diff []byte) []byte {
	return filterChartOnlyAdditions(diff, nil, nil)
}

// filterChartOnlyAdditions keeps pure-addition hunks. When base and theirs
// are non-nil, hunks whose added lines already occur contiguously in theirs
// but not in base are dropped. The result always uses a timestamp-free
// header and a trailing newline so GNU patch accepts it.
func filterChartOnlyAdditions(diff, base, theirs []byte) []byte {
	lines := strings.Split(string(diff), "\n")
	// A trailing newline becomes a final empty element. Leaving it on the
	// last hunk emits a blank diff line, which GNU patch rejects as
	// "malformed patch".
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var hunks [][]string
	var cur []string
	var curHasDeletion bool

	flush := func() {
		if len(cur) == 0 {
			return
		}
		h := cur
		deleted := curHasDeletion
		cur = nil
		curHasDeletion = false
		if deleted {
			return
		}
		var body []string
		var added []string
		for _, line := range h[1:] {
			if line == "" {
				continue
			}
			body = append(body, line)
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				added = append(added, line[1:])
			}
		}
		if len(added) == 0 {
			return
		}
		if base == nil || theirs == nil {
			hunks = append(hunks, append([]string{h[0]}, body...))
			return
		}
		// A pure-addition hunk can mix chart-only lines (Helm templating)
		// with an upstream addition the chart already contains. Replay only
		// the chart-only runs; the upstream lines arrive via the 3-way merge.
		oldStart, newStart := hunkStarts(h[0])
		for _, run := range chartOnlyRuns(added, base, theirs) {
			hunk := []string{fmt.Sprintf("@@ -%d,0 +%d,%d @@", oldStart, newStart, len(run))}
			for _, line := range run {
				hunk = append(hunk, "+"+line)
			}
			hunks = append(hunks, hunk)
			newStart += len(run)
		}
	}

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") ||
			strings.HasPrefix(line, "diff --git ") || strings.HasPrefix(line, "index "):
			continue
		case strings.HasPrefix(line, "@@"):
			flush()
			cur = []string{line}
		case len(cur) > 0:
			if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "--- ") {
				curHasDeletion = true
			}
			cur = append(cur, line)
		}
	}
	flush()

	if len(hunks) == 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString("--- a\n+++ b\n")
	for _, h := range hunks {
		for _, line := range h {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	return []byte(b.String())
}

// hunkStarts reads the old and new start lines from a unified hunk header.
func hunkStarts(header string) (oldStart, newStart int) {
	// @@ -35,0 +39,3 @@
	rest := strings.TrimPrefix(header, "@@ ")
	oldStart = leadingInt(strings.TrimPrefix(rest, "-"))
	if i := strings.Index(rest, "+"); i >= 0 {
		newStart = leadingInt(rest[i+1:])
	}
	if oldStart == 0 {
		oldStart = 1
	}
	if newStart == 0 {
		newStart = oldStart
	}
	return oldStart, newStart
}

func leadingInt(s string) int {
	n := 0
	saw := false
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		saw = true
		n = n*10 + int(c-'0')
	}
	if !saw {
		return 0
	}
	return n
}

// chartOnlyRuns drops contiguous slices of added lines that already exist in
// theirs and not in base. The remaining slices are chart-only additions.
func chartOnlyRuns(added []string, base, theirs []byte) [][]string {
	var runs [][]string
	var cur []string
	for i := 0; i < len(added); {
		if n := longestUpstreamPrefix(added[i:], base, theirs); n > 0 {
			if len(cur) > 0 {
				runs = append(runs, cur)
				cur = nil
			}
			i += n
			continue
		}
		cur = append(cur, added[i])
		i++
	}
	if len(cur) > 0 {
		runs = append(runs, cur)
	}
	return runs
}

// longestUpstreamPrefix returns the longest prefix of added that is a
// contiguous line sequence in theirs and not in base.
func longestUpstreamPrefix(added []string, base, theirs []byte) int {
	best := 0
	for k := 1; k <= len(added); k++ {
		if linesAppearContiguous(theirs, added[:k]) && !linesAppearContiguous(base, added[:k]) {
			best = k
		}
	}
	return best
}

// linesAppearContiguous reports whether lines occur in order, adjacent, in file.
func linesAppearContiguous(file []byte, lines []string) bool {
	if len(lines) == 0 {
		return false
	}
	fileLines := strings.Split(string(file), "\n")
	for i := 0; i+len(lines) <= len(fileLines); i++ {
		match := true
		for j, line := range lines {
			if fileLines[i+j] != line {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
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

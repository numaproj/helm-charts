package mirror

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/numaproj/helm-charts/upgrade/common"
	"helm.sh/helm/v3/pkg/chartutil"
)

// SyncOptions controls a single `sync` run.
type SyncOptions struct {
	// FromVersion is the previous numaflow version the chart was synced
	// against. If empty, it defaults to the appVersion recorded in the
	// committed (HEAD) Chart.yaml, prefixed with `v`. HEAD is used rather than
	// the working tree because the documented release flow runs
	// `upgrade-charts` first, which rewrites the working-tree appVersion to the
	// new version.
	FromVersion string

	// ToVersion is the new numaflow version to sync to. Required.
	ToVersion string

	// Apply, when true, writes clean merges back to the chart files. When
	// false, all merges are written to RejectsDir for human review only.
	Apply bool

	// Only, when non-empty, restricts processing to mirror files whose
	// LocalPath basename is in the set. Useful for incremental sync runs.
	Only []string

	// BaseDir is the chart root (e.g. /path/to/helm-charts/charts/numaflow/).
	// If empty, common.BaseDir is used.
	BaseDir string

	// RejectsDir is where conflict outputs and (when !Apply) clean-merge
	// outputs are written. If empty, defaults to <BaseDir>/../upgrade/.merge-rejects.
	RejectsDir string

	// Out is the log destination. If nil, defaults to os.Stdout.
	Out io.Writer
}

// FileReport describes the result for a single mirrored file.
type FileReport struct {
	LocalPath    string
	UpstreamPath string
	Status       string // see ReportStatus* constants
	Detail       string
}

// Report status strings (kept as plain strings so they appear cleanly in the
// summary table without further mapping).
const (
	ReportStatusNoChange = "no-change"
	ReportStatusApplied  = "applied"
	ReportStatusReady    = "ready (not applied)"
	ReportStatusConflict = "conflict"
	ReportStatusMissing  = "upstream-missing"
	ReportStatusError    = "error"
)

// Report aggregates per-file results plus counters used to decide exit code.
type Report struct {
	Files     []FileReport
	NoChange  int
	Applied   int
	Ready     int
	Conflicts int
	Missing   int
	Errors    int
}

// HasFailures reports whether the run should be considered a failure for
// the purposes of exit code. Conflicts and errors are failures; upstream-
// missing files are surfaced but not failures (they require maintainer
// judgement but are not a tool bug).
func (r Report) HasFailures() bool {
	return r.Conflicts > 0 || r.Errors > 0
}

// Run executes a sync pass over the registry according to opts. It returns
// a Report describing per-file outcomes; the error is non-nil only for
// pre-flight failures (missing tools, missing version, IO error setting up
// the workspace) — per-file outcomes are reported via the Report.
func Run(opts SyncOptions) (Report, error) {
	if opts.ToVersion == "" {
		return Report{}, fmt.Errorf("ToVersion is required (pass --to-version or set NUMAFLOW_VERSION)")
	}
	if opts.BaseDir == "" {
		opts.BaseDir = common.BaseDir
	}
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.RejectsDir == "" {
		opts.RejectsDir = filepath.Join(filepath.Dir(strings.TrimRight(opts.BaseDir, "/")), "..", "upgrade", ".merge-rejects")
	}

	if err := EnsureToolsAvailable(); err != nil {
		return Report{}, err
	}

	// Resolve FromVersion default from the committed (HEAD) Chart.yaml.
	//
	// We deliberately read HEAD rather than the working tree: the documented
	// release flow runs `upgrade-charts` first, which rewrites the working-tree
	// appVersion to the *new* version. Reading that back would make from==to
	// and silently turn the whole sync into a no-op. HEAD still carries the
	// previous release's appVersion.
	if opts.FromVersion == "" {
		appVersion, err := appVersionFromGitHEAD(opts.BaseDir)
		if err != nil {
			return Report{}, fmt.Errorf("resolve default --from-version from git HEAD: %w; pass --from-version explicitly", err)
		}
		if appVersion == "" {
			return Report{}, fmt.Errorf("HEAD Chart.yaml has empty appVersion; pass --from-version explicitly")
		}
		opts.FromVersion = "v" + strings.TrimPrefix(appVersion, "v")
		fmt.Fprintf(opts.Out, "info: --from-version not set; using HEAD Chart.yaml appVersion %s\n", opts.FromVersion)
	}

	// Normalise both versions to a leading `v` so the guard below and the
	// downstream upstream URLs are consistent regardless of how the caller
	// spelled them.
	opts.FromVersion = "v" + strings.TrimPrefix(opts.FromVersion, "v")
	opts.ToVersion = "v" + strings.TrimPrefix(opts.ToVersion, "v")

	// Guard: from == to means there is nothing to diff. Most commonly this
	// happens when the `upgrade-charts` bump has already been committed (so
	// HEAD carries the new version too), or --from-version was set equal to
	// --to-version by hand. Fail loudly rather than reporting every file as
	// no-change.
	if opts.FromVersion == opts.ToVersion {
		return Report{}, fmt.Errorf("--from-version (%s) equals --to-version (%s); nothing to compare — pass an explicit --from-version for the previous release", opts.FromVersion, opts.ToVersion)
	}

	// Pre-flight: confirm the to-version tag exists upstream so we fail
	// fast on a typo.
	if err := verifyVersionExists(opts.ToVersion); err != nil {
		return Report{}, fmt.Errorf("to-version: %w", err)
	}

	if err := os.MkdirAll(opts.RejectsDir, 0o755); err != nil {
		return Report{}, fmt.Errorf("create rejects dir: %w", err)
	}

	files := filterRegistry(MirroredFiles, opts.Only)
	if len(files) == 0 {
		return Report{}, fmt.Errorf("no mirrored files match --only filter %v", opts.Only)
	}

	fmt.Fprintf(opts.Out, "syncing %d mirrored files from %s -> %s (apply=%v)\n",
		len(files), opts.FromVersion, opts.ToVersion, opts.Apply)

	var report Report
	for _, m := range files {
		fr := processOneFile(opts, m)
		report.Files = append(report.Files, fr)
		switch fr.Status {
		case ReportStatusNoChange:
			report.NoChange++
		case ReportStatusApplied:
			report.Applied++
		case ReportStatusReady:
			report.Ready++
		case ReportStatusConflict:
			report.Conflicts++
		case ReportStatusMissing:
			report.Missing++
		case ReportStatusError:
			report.Errors++
		}
	}

	writeSummary(opts.Out, report, opts)
	return report, nil
}

func filterRegistry(all []MirroredFile, only []string) []MirroredFile {
	if len(only) == 0 {
		return all
	}
	want := map[string]bool{}
	for _, name := range only {
		want[name] = true
	}
	var out []MirroredFile
	for _, m := range all {
		if want[filepath.Base(m.LocalPath)] {
			out = append(out, m)
		}
	}
	return out
}

// processOneFile runs L1 then L2 for a single mirrored file and produces a
// FileReport. It writes any conflict/merge output to RejectsDir and, if
// Apply is set and the merge was clean, also writes the merged content back
// to the chart file.
func processOneFile(opts SyncOptions, m MirroredFile) FileReport {
	rep := FileReport{LocalPath: m.LocalPath, UpstreamPath: m.UpstreamPath}

	up := FetchUpstreamPair(m, opts.FromVersion, opts.ToVersion)
	switch up.Status {
	case StatusNoChange:
		rep.Status = ReportStatusNoChange
		return rep
	case StatusUpstreamMissing:
		rep.Status = ReportStatusMissing
		rep.Detail = up.Err.Error()
		fmt.Fprintf(opts.Out, "  %s: upstream missing — %v\n", m.LocalPath, up.Err)
		return rep
	case StatusError:
		rep.Status = ReportStatusError
		rep.Detail = up.Err.Error()
		fmt.Fprintf(opts.Out, "  %s: error fetching upstream — %v\n", m.LocalPath, up.Err)
		return rep
	case StatusChanged:
		// fall through to merge
	}

	chartPath := filepath.Join(opts.BaseDir, m.LocalPath)
	chartCopy, err := os.ReadFile(chartPath)
	if err != nil {
		rep.Status = ReportStatusError
		rep.Detail = err.Error()
		fmt.Fprintf(opts.Out, "  %s: read chart file — %v\n", m.LocalPath, err)
		return rep
	}

	merge := ThreeWayMerge(chartCopy, up.BaseBlob, up.HeadBlob)
	switch merge.Outcome {
	case MergeClean:
		if opts.Apply {
			if err := os.WriteFile(chartPath, merge.Merged, 0o644); err != nil {
				rep.Status = ReportStatusError
				rep.Detail = err.Error()
				fmt.Fprintf(opts.Out, "  %s: write applied merge — %v\n", m.LocalPath, err)
				return rep
			}
			rep.Status = ReportStatusApplied
			fmt.Fprintf(opts.Out, "  %s: clean merge applied\n", m.LocalPath)
		} else {
			out := filepath.Join(opts.RejectsDir, filepath.Base(m.LocalPath)+".merged")
			if err := os.WriteFile(out, merge.Merged, 0o644); err != nil {
				rep.Status = ReportStatusError
				rep.Detail = err.Error()
				fmt.Fprintf(opts.Out, "  %s: write merged output — %v\n", m.LocalPath, err)
				return rep
			}
			rep.Status = ReportStatusReady
			rep.Detail = out
			fmt.Fprintf(opts.Out, "  %s: clean merge ready (not applied) — %s\n", m.LocalPath, out)
		}
	case MergeConflict:
		out := filepath.Join(opts.RejectsDir, filepath.Base(m.LocalPath)+".conflict")
		if err := os.WriteFile(out, merge.Merged, 0o644); err != nil {
			rep.Status = ReportStatusError
			rep.Detail = err.Error()
			fmt.Fprintf(opts.Out, "  %s: write conflict output — %v\n", m.LocalPath, err)
			return rep
		}
		rep.Status = ReportStatusConflict
		rep.Detail = fmt.Sprintf("%d conflict region(s) -> %s", merge.NumConflicts, out)
		fmt.Fprintf(opts.Out, "  %s: %s\n", m.LocalPath, rep.Detail)
	case MergeErrored:
		rep.Status = ReportStatusError
		rep.Detail = merge.Err.Error()
		fmt.Fprintf(opts.Out, "  %s: merge error — %v\n", m.LocalPath, merge.Err)
	}
	return rep
}

// appVersionFromGitHEAD returns the appVersion recorded in the committed
// (HEAD) copy of charts/numaflow/Chart.yaml — i.e. the value before any
// working-tree bump by `upgrade-charts`. baseDir is used only as the git
// working directory; the path passed to `git show` is repo-root relative so
// the lookup works regardless of where the binary was invoked from.
//
// git availability is guaranteed by EnsureToolsAvailable, which Run calls
// before reaching here.
func appVersionFromGitHEAD(baseDir string) (string, error) {
	rel := "charts/numaflow/" + chartutil.ChartfileName
	cmd := exec.Command("git", "show", "HEAD:"+rel)
	cmd.Dir = baseDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git show HEAD:%s: %w (stderr: %s)", rel, err, strings.TrimSpace(stderr.String()))
	}

	// Reuse chartutil's loader (via a temp file) so parsing semantics match
	// the working-tree path exactly.
	tmp, err := os.CreateTemp("", "chart-head-*.yaml")
	if err != nil {
		return "", fmt.Errorf("create temp chart file: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(out); err != nil {
		tmp.Close()
		return "", fmt.Errorf("write temp chart file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close temp chart file: %w", err)
	}
	chart, err := chartutil.LoadChartfile(tmp.Name())
	if err != nil {
		return "", fmt.Errorf("parse HEAD Chart.yaml: %w", err)
	}
	return chart.AppVersion, nil
}

// verifyVersionExists checks the GitHub releases endpoint for the given tag.
// This duplicates internal.IsVersionExists in spirit but keeps the mirror
// package free of an internal/internal import path quirk.
func verifyVersionExists(version string) error {
	url := fmt.Sprintf("https://api.github.com/repos/numaproj/numaflow/releases/tags/%s", version)
	_, err := common.Download(url)
	if err != nil && strings.Contains(err.Error(), "404") {
		return fmt.Errorf("numaflow tag %s does not exist on github.com/numaproj/numaflow", version)
	}
	return nil
}

// writeSummary prints a tidy markdown-friendly summary at the end of a run.
func writeSummary(out io.Writer, r Report, opts SyncOptions) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Mirror sync summary")
	fmt.Fprintln(out, "===================")
	fmt.Fprintf(out, "  no-change:        %d\n", r.NoChange)
	fmt.Fprintf(out, "  applied:          %d\n", r.Applied)
	fmt.Fprintf(out, "  ready:            %d\n", r.Ready)
	fmt.Fprintf(out, "  conflict:         %d\n", r.Conflicts)
	fmt.Fprintf(out, "  upstream-missing: %d\n", r.Missing)
	fmt.Fprintf(out, "  errors:           %d\n", r.Errors)
	fmt.Fprintln(out)

	// Per-file table, sorted by LocalPath for stable output.
	sorted := make([]FileReport, len(r.Files))
	copy(sorted, r.Files)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].LocalPath < sorted[j].LocalPath })
	for _, f := range sorted {
		line := fmt.Sprintf("  %-20s %s", f.Status, f.LocalPath)
		if f.Detail != "" && f.Status != ReportStatusNoChange {
			line += "    " + f.Detail
		}
		fmt.Fprintln(out, line)
	}

	if r.HasFailures() {
		fmt.Fprintf(out, "\nresult: FAILED (%d conflict, %d error) — review %s\n", r.Conflicts, r.Errors, opts.RejectsDir)
	} else if !opts.Apply && r.Ready > 0 {
		fmt.Fprintf(out, "\nresult: OK — review proposed merges under %s; re-run with --apply to write them back to the chart\n", opts.RejectsDir)
	} else {
		fmt.Fprintln(out, "\nresult: OK")
	}
}

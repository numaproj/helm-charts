package mirror

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/numaproj/helm-charts/upgrade/common"
	"helm.sh/helm/v3/pkg/chartutil"
)

// SyncOptions controls a single `sync` run.
type SyncOptions struct {
	// FromVersion is the previous numaflow version the chart was synced
	// against. If empty, it defaults to the chart's Chart.yaml appVersion
	// (prefixed with `v`).
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
	ReportStatusNoChange    = "no-change"
	ReportStatusApplied     = "applied"
	ReportStatusReady       = "ready (not applied)"
	ReportStatusConflict    = "conflict"
	ReportStatusMissing     = "upstream-missing"
	ReportStatusError       = "error"
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

	// Resolve FromVersion default from Chart.yaml.
	if opts.FromVersion == "" {
		chartFile := filepath.Join(opts.BaseDir, chartutil.ChartfileName)
		chart, err := chartutil.LoadChartfile(chartFile)
		if err != nil {
			return Report{}, fmt.Errorf("read %s: %w", chartFile, err)
		}
		if chart.AppVersion == "" {
			return Report{}, fmt.Errorf("Chart.yaml has empty appVersion; pass --from-version explicitly")
		}
		opts.FromVersion = "v" + strings.TrimPrefix(chart.AppVersion, "v")
		fmt.Fprintf(opts.Out, "info: --from-version not set; using Chart.yaml appVersion %s\n", opts.FromVersion)
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

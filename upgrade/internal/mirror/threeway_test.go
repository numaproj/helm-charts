package mirror

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func requireTools(t *testing.T) {
	t.Helper()
	for _, tool := range []string{"git", "diff", "patch"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not on PATH: %v", tool, err)
		}
	}
}

func TestThreeWayMerge_CleanMerge_UpstreamAddsKey(t *testing.T) {
	// Upstream adds a new key adjacent to the chart's labels block. The
	// chart-only addition pre-processing should let this merge cleanly.
	requireTools(t)

	base := []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cmp\ndata:\n  foo: \"1\"\n")
	theirs := []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cmp\ndata:\n  foo: \"1\"\n  bar: \"2\"\n")
	chart := []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cmp\n  labels:\n    {{- include \"numaflow.labels\" . | nindent 4 }}\n  namespace: {{ .Release.Namespace }}\ndata:\n  foo: \"1\"\n")

	res := ThreeWayMerge(chart, base, theirs)
	if res.Outcome != MergeClean {
		t.Fatalf("want MergeClean, got %s (err: %v)\nmerged:\n%s", res.Outcome, res.Err, res.Merged)
	}
	if !bytes.Contains(res.Merged, []byte("bar: \"2\"")) {
		t.Fatalf("expected new key bar in merged output:\n%s", res.Merged)
	}
	if !bytes.Contains(res.Merged, []byte("{{- include \"numaflow.labels\"")) {
		t.Fatalf("chart-only labels block missing from merged output:\n%s", res.Merged)
	}
	if !bytes.Contains(res.Merged, []byte("{{ .Release.Namespace }}")) {
		t.Fatalf("chart-only namespace line missing from merged output:\n%s", res.Merged)
	}
}

func TestThreeWayMerge_Conflict_BothChangedSameLine(t *testing.T) {
	// Chart templated a value that upstream is now changing to a new
	// hardcoded value. Git's 3-way merge should detect this as a real
	// conflict and produce inline conflict markers.
	requireTools(t)

	base := []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cmp\ndata:\n  foo: \"1\"\n  unrelated: \"x\"\n")
	theirs := []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cmp\ndata:\n  foo: \"2\"\n  unrelated: \"x\"\n")
	chart := []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cmp\ndata:\n  foo: {{ .Values.foo | quote }}\n  unrelated: \"x\"\n")

	res := ThreeWayMerge(chart, base, theirs)
	if res.Outcome != MergeConflict {
		t.Fatalf("want MergeConflict, got %s\nmerged:\n%s", res.Outcome, res.Merged)
	}
	if res.NumConflicts < 1 {
		t.Fatalf("expected at least 1 conflict, got %d", res.NumConflicts)
	}
	merged := string(res.Merged)
	for _, marker := range []string{"<<<<<<<", "|||||||", "=======", ">>>>>>>"} {
		if !strings.Contains(merged, marker) {
			t.Fatalf("expected conflict marker %q in merged output:\n%s", marker, merged)
		}
	}
	if !strings.Contains(merged, "{{ .Values.foo | quote }}") {
		t.Fatalf("chart template lost from conflict region:\n%s", merged)
	}
	if !strings.Contains(merged, "foo: \"2\"") {
		t.Fatalf("upstream's value missing from conflict region:\n%s", merged)
	}
}

func TestThreeWayMerge_TemplatePassthrough(t *testing.T) {
	// Upstream changes a non-templated line; the chart's templated lines and
	// block directives should survive the merge untouched.
	//
	// The fixture deliberately places the chart's modified line and upstream's
	// modified line several rows apart, outside git xdiff's default context
	// window. This reflects real chart files (chart additions sit in the
	// metadata block; upstream tweaks sit in the data block) and avoids the
	// known limitation where two non-overlapping line modifications within a
	// 3-line window get bundled into one hunk and flagged as a conflict.
	requireTools(t)

	base := []byte(
		"data:\n" +
			"  foo: \"1\"\n" +
			"  alpha: \"a\"\n" +
			"  beta: \"b\"\n" +
			"  gamma: \"g\"\n" +
			"  shared: \"old\"\n")
	theirs := []byte(
		"data:\n" +
			"  foo: \"1\"\n" +
			"  alpha: \"a\"\n" +
			"  beta: \"b\"\n" +
			"  gamma: \"g\"\n" +
			"  shared: \"new\"\n")
	chart := []byte(
		"data:\n" +
			"  foo: {{ .Values.foo | quote }}\n" +
			"  alpha: \"a\"\n" +
			"  beta: \"b\"\n" +
			"  gamma: \"g\"\n" +
			"  shared: \"old\"\n" +
			"  {{- if .Values.extra.enabled }}\n" +
			"  extra: \"yes\"\n" +
			"  {{- end }}\n")

	res := ThreeWayMerge(chart, base, theirs)
	if res.Outcome != MergeClean {
		t.Fatalf("want MergeClean, got %s (err: %v)\nmerged:\n%s", res.Outcome, res.Err, res.Merged)
	}
	merged := string(res.Merged)
	if !strings.Contains(merged, "shared: \"new\"") {
		t.Fatalf("upstream change to `shared` missing:\n%s", merged)
	}
	if !strings.Contains(merged, "{{ .Values.foo | quote }}") {
		t.Fatalf("inline template lost:\n%s", merged)
	}
	if !strings.Contains(merged, "{{- if .Values.extra.enabled }}") || !strings.Contains(merged, "{{- end }}") {
		t.Fatalf("block directives lost:\n%s", merged)
	}
}

func TestThreeWayMerge_AdjacencyLimitation(t *testing.T) {
	// Documents a known limitation: when the chart's modification of an
	// upstream line and upstream's modification of a different upstream line
	// sit within git xdiff's context window (~3 lines), they are reported as
	// a single conflicting hunk even though the lines do not overlap.
	//
	// This test pins that behaviour so a future change to the merge strategy
	// either preserves it intentionally or is caught for review.
	requireTools(t)

	base := []byte("data:\n  foo: \"1\"\n  shared: \"old\"\n")
	theirs := []byte("data:\n  foo: \"1\"\n  shared: \"new\"\n")
	chart := []byte("data:\n  foo: {{ .Values.foo | quote }}\n  shared: \"old\"\n")

	res := ThreeWayMerge(chart, base, theirs)
	if res.Outcome != MergeConflict {
		t.Skipf("adjacency no longer produces a false-positive conflict (improvement, not a regression): %s", res.Outcome)
	}
	if !strings.Contains(string(res.Merged), "{{ .Values.foo | quote }}") {
		t.Fatalf("chart template lost in adjacency conflict:\n%s", res.Merged)
	}
}

func TestThreeWayMerge_NoUpstreamChange_NoOp(t *testing.T) {
	// When base == theirs, the chart copy should pass through unchanged.
	requireTools(t)

	base := []byte("data:\n  foo: \"1\"\n")
	chart := []byte("data:\n  foo: {{ .Values.foo | quote }}\n")

	res := ThreeWayMerge(chart, base, base)
	if res.Outcome != MergeClean {
		t.Fatalf("want MergeClean, got %s", res.Outcome)
	}
	if !bytes.Equal(res.Merged, chart) {
		t.Fatalf("no-op merge altered chart copy\nwant: %s\ngot:  %s", chart, res.Merged)
	}
}

func TestFilterAdditionsOnly(t *testing.T) {
	// Confirm the diff filter keeps pure-addition hunks and drops hunks with
	// any deletion line.
	diff := []byte(
		"--- a/base.in\n" +
			"+++ b/ours.in\n" +
			"@@ -1,2 +1,3 @@\n" +
			" first\n" +
			"+inserted-only\n" +
			" second\n" +
			"@@ -10,2 +11,2 @@\n" +
			"-deleted\n" +
			"+replaced\n" +
			" tail\n")
	out := filterAdditionsOnly(diff)
	if !bytes.Contains(out, []byte("inserted-only")) {
		t.Errorf("addition hunk dropped:\n%s", out)
	}
	if bytes.Contains(out, []byte("deleted")) {
		t.Errorf("modification hunk kept:\n%s", out)
	}
	if bytes.Contains(out, []byte("replaced")) {
		t.Errorf("modification hunk kept:\n%s", out)
	}
}

package mirror

import (
	"crypto/sha1"
	"encoding/hex"
	"regexp"
	"strings"
)

// Tokenize replaces every Helm template construct in src with a deterministic
// sentinel that is safe to round-trip through `git merge-file`. It returns the
// tokenized bytes and a map from sentinel -> original text for use by
// Detokenize.
//
// Two forms are recognised:
//
//  1. Inline expressions like `key: {{ .Values.X | quote }}`. The `{{...}}`
//     run is replaced with `__NFTPLEXPR_<sha1[:8]>__`. The map stores just the
//     expression text.
//
//  2. Lines whose non-whitespace content is entirely template directives, e.g.
//     `{{- if X }}`, `{{- end }}`, `{{- include "labels" . | nindent 4 }}`.
//     The whole line is replaced with `<indent># __NFTPLLINE_<sha1[:8]>__` so
//     that the tokenized YAML remains parseable. The map stores the entire
//     original line (without its trailing newline).
//
// Substituting templates with sentinels means git's line-based merge sees
// stable identifiers on both the chart and upstream sides for unchanged
// templated lines, so it only conflicts where upstream and chart genuinely
// disagree on the same line.
const (
	sentinelInlinePrefix = "__NFTPLEXPR_"
	sentinelBlockPrefix  = "__NFTPLLINE_"
	sentinelSuffix       = "__"
)

var (
	exprRe   = regexp.MustCompile(`\{\{-?[\s\S]*?-?\}\}`)
	blockRe  = regexp.MustCompile(`^([ \t]*)# (` + regexp.QuoteMeta(sentinelBlockPrefix) + `[0-9a-f]+` + regexp.QuoteMeta(sentinelSuffix) + `)\s*$`)
	inlineRe = regexp.MustCompile(regexp.QuoteMeta(sentinelInlinePrefix) + `[0-9a-f]+` + regexp.QuoteMeta(sentinelSuffix))
)

// Tokenize converts a templated chart file into a sentinel-substituted form
// suitable for `git merge-file`. The returned map records the substitutions
// so Detokenize can restore the original content exactly.
func Tokenize(src []byte) ([]byte, map[string]string) {
	subs := map[string]string{}
	lines := strings.Split(string(src), "\n")
	for i, line := range lines {
		// Block form: the entire non-whitespace content of the line is template
		// directives. Replace the whole line with a YAML comment sentinel.
		stripped := exprRe.ReplaceAllString(line, "")
		if strings.Contains(line, "{{") && strings.TrimSpace(stripped) == "" {
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			h := hash(line)
			sentinel := sentinelBlockPrefix + h + sentinelSuffix
			subs[sentinel] = line
			lines[i] = indent + "# " + sentinel
			continue
		}
		// Inline form: replace each `{{...}}` run individually.
		lines[i] = exprRe.ReplaceAllStringFunc(line, func(match string) string {
			h := hash(match)
			sentinel := sentinelInlinePrefix + h + sentinelSuffix
			subs[sentinel] = match
			return sentinel
		})
	}
	return []byte(strings.Join(lines, "\n")), subs
}

// Detokenize reverses Tokenize: it replaces every sentinel in src with the
// original text recorded in subs.
//
// Block sentinels are detected at line level so they restore the original line
// verbatim, including the original indent and any trailing whitespace.
// Inline sentinels are replaced as substrings within the surviving line.
//
// Unknown sentinels (not present in subs — for example, if the merge inserted
// sentinel text from the other side via a hunk we did not write) are left
// untouched so the caller can flag them.
func Detokenize(src []byte, subs map[string]string) []byte {
	lines := strings.Split(string(src), "\n")
	for i, line := range lines {
		if m := blockRe.FindStringSubmatch(line); m != nil {
			if original, ok := subs[m[2]]; ok {
				lines[i] = original
				continue
			}
		}
		lines[i] = inlineRe.ReplaceAllStringFunc(line, func(sentinel string) string {
			if original, ok := subs[sentinel]; ok {
				return original
			}
			return sentinel
		})
	}
	return []byte(strings.Join(lines, "\n"))
}

func hash(s string) string {
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:4])
}

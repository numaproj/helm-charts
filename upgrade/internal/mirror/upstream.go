package mirror

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/numaproj/helm-charts/upgrade/common"
)

// Status is the outcome of fetching the upstream pair for a single mirrored
// file. It tells the orchestrator whether to run a 3-way merge, skip the
// file, or surface an error.
type Status int

const (
	StatusUnknown Status = iota
	// StatusNoChange means upstream@vOld and upstream@vNew are byte-identical
	// (after normalisation). The chart copy needs no update.
	StatusNoChange
	// StatusChanged means upstream@vOld and upstream@vNew differ. The chart
	// copy should be brought up to date via a 3-way merge.
	StatusChanged
	// StatusUpstreamMissing means upstream does not have the file at one or
	// both versions (404). Most commonly: a file that was added in a later
	// upstream release, or a path that was renamed. The maintainer must
	// resolve manually.
	StatusUpstreamMissing
	// StatusError means the fetch failed for an unexpected reason (network,
	// rate limit exhaustion, IO).
	StatusError
)

func (s Status) String() string {
	switch s {
	case StatusNoChange:
		return "no-change"
	case StatusChanged:
		return "changed"
	case StatusUpstreamMissing:
		return "upstream-missing"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// UpstreamResult is the L1 output for a single mirrored file.
type UpstreamResult struct {
	Mirror   MirroredFile
	Status   Status
	BaseBlob []byte // upstream@vOld, populated when Status == StatusChanged
	HeadBlob []byte // upstream@vNew, populated when Status == StatusChanged
	Err      error
}

// FetchUpstreamPair downloads the upstream blobs for vOld and vNew, normalises
// line endings, and reports whether the upstream changed. This is L1 of the
// design: the primary signal for whether any work is needed on a file.
func FetchUpstreamPair(m MirroredFile, vOld, vNew string) UpstreamResult {
	res := UpstreamResult{Mirror: m}

	baseURL := common.GithubBaseURL + vOld + m.UpstreamPath
	headURL := common.GithubBaseURL + vNew + m.UpstreamPath

	baseStr, err := common.Download(baseURL)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			res.Status = StatusUpstreamMissing
			res.Err = fmt.Errorf("upstream@%s not found at %s", vOld, m.UpstreamPath)
			return res
		}
		res.Status = StatusError
		res.Err = fmt.Errorf("fetch %s: %w", baseURL, err)
		return res
	}

	headStr, err := common.Download(headURL)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			res.Status = StatusUpstreamMissing
			res.Err = fmt.Errorf("upstream@%s not found at %s", vNew, m.UpstreamPath)
			return res
		}
		res.Status = StatusError
		res.Err = fmt.Errorf("fetch %s: %w", headURL, err)
		return res
	}

	base := normalizeBlob([]byte(baseStr))
	head := normalizeBlob([]byte(headStr))

	if bytes.Equal(base, head) {
		res.Status = StatusNoChange
		return res
	}

	res.Status = StatusChanged
	res.BaseBlob = base
	res.HeadBlob = head
	return res
}

// normalizeBlob strips a UTF-8 BOM and converts CRLF line endings to LF so
// byte equality reflects content equality rather than incidental encoding
// differences.
func normalizeBlob(b []byte) []byte {
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	return b
}

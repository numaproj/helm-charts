package common

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Download fetches the contents of url, retrying on HTTP 429. It returns the
// response body on success and the last error observed after all retries on
// failure. Non-429 errors (including 404) fail fast on the first attempt.
//
// Retry timing is preserved verbatim from the previous downloadFileDataWithRetry
// implementation to avoid changing observable behaviour while extracting the
// helper.
func Download(url string) (string, error) {
	const maxRetries = 3
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		data, err := DownloadOnce(url)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if !strings.Contains(err.Error(), "429") {
			break
		}
		time.Sleep(time.Duration(2 << (5 * i)))
	}
	return "", fmt.Errorf("failed to download %s after retries: %w", url, lastErr)
}

// DownloadOnce performs a single GET against url with no retry logic. It is
// exported so callers that need to inspect status codes themselves (e.g.
// IsVersionExists, which expects a 404 to mean "no such tag") can use it
// directly.
func DownloadOnce(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP Request Failed with Status: %d %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response data: %w", err)
	}

	return string(data), nil
}

package api

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/numaproj/helm-charts/upgrade/common"
)

// Retry DownloadFileData retries the download of file data from the given URL if failed with 429 error code
// and returns the file data as a string.

func DownloadFileDataWithRetry(url string) (string, error) {
	const maxRetries = 3
	var err error
	var data string

	for i := 0; i < maxRetries; i++ {
		data, err = DownloadFileData(url)
		if err == nil {
			return data, nil
		}
		if err.Error() == "HTTP Request Failed with Status: 429 Too Many Requests" {
			time.Sleep(2 << (5 * i)) // Exponential backoff
			continue                 // Retry on 429 error
		}
		break // Break on other errors
	}

	return "", fmt.Errorf("failed to download file data after %d attempts: %w", maxRetries, err)
}

func DownloadFileData(url string) (string, error) {
	resp, err := http.Get(common.GithubBaseURL + url)
	if err != nil {
		return "", fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	// Check if the HTTP request was successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP Request Failed with Status: %d %s", resp.StatusCode, resp.Status)
	}

	// Read the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response data: %w", err)
	}

	return string(data), nil
}

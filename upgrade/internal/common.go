package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/numaproj/helm-charts/upgrade/common"
)

// UpdateChartFile updates the charts files data with the data fetched from upstream
func updateFiles(localFilePath, url, numaflowVersion string, namespaced bool) error {
	yamlContent, err := os.ReadFile(localFilePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	lines := strings.Split(string(yamlContent), "\n")
	// remove all blank lines from the end of the file
	// Remove blank lines from the end of the lines slice
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	var firstLine, lastLine string
	if len(lines) > 0 {
		firstLine = lines[0]
		lastLine = lines[len(lines)-1]
	} else {
		firstLine = common.ContentSeparator
		lastLine = common.ContentSeparator
	}

	latestData, err := downloadFileDataWithRetry(common.GithubBaseURL + numaflowVersion + url)
	if err != nil {
		return fmt.Errorf("error fetching latest data for file: %s, err:%v", localFilePath, err)
	}

	// Update labels and context separator in file
	updatedDataLines := strings.Split(latestData, "\n")
	updatedDataLines = append([]string{firstLine}, updatedDataLines...)
	updatedDataLines = append(updatedDataLines, lastLine)
	updatedDataLines = addLabelToData(updatedDataLines)
	if namespaced {
		updatedDataLines = addNamespace(updatedDataLines)
	}

	updatedContent := strings.Join(updatedDataLines, "\n")

	err = os.WriteFile(localFilePath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing back to file: %s, err:%v", localFilePath, err)
	}

	return nil
}

// addLabelToData adds the default label to the data
func addLabelToData(data []string) []string {
	for i, line := range data {
		if strings.Contains(line, "labels:") {
			data[i] = line + "\n" + common.DefaultLabel
			return data
		}
	}
	for i, line := range data {
		if strings.Contains(line, "metadata:") {
			// Add formatted space before the labels
			line = strings.Replace(line, "metadata:", "metadata:\n  labels:", 1)
			line = line + "\n" + common.DefaultLabel
			data[i] = line
		}
	}

	return data
}

// addNamespace adds the namespace to the data
func addNamespace(data []string) []string {
	for i, line := range data {
		if strings.Contains(line, "metadata:") {
			line = line + "\n" + "  namespace: {{ .Release.Namespace }}"
			data[i] = line
		}
	}

	return data
}

// IsVersionExists checks if the version exists in the GitHub releases
func IsVersionExists(numaflowVersion string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/numaproj/numaflow/releases/tags/%s", numaflowVersion)
	_, err := downloadFileDataWithRetry(url)
	if err != nil && strings.Contains(err.Error(), "404") {
		return false, err
	}

	return true, nil
}

// DownloadFileData downloads the file data from the given URL
func DownloadFileData(url string) (string, error) {
	resp, err := http.Get(url)
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

// downloadFileDataWithRetry downloads the file data with retry logic
func downloadFileDataWithRetry(url string) (string, error) {
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

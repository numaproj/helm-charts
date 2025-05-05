package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/numaproj/helm-charts/upgrade/api"
	"github.com/numaproj/helm-charts/upgrade/common"
)

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

	latestData, err := api.DownloadFileDataWithRetry(common.GithubBaseURL + numaflowVersion + url)
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

func addNamespace(data []string) []string {
	for i, line := range data {
		if strings.Contains(line, "metadata:") {
			// Add formatted space before the namespace
			//line = strings.Replace(line, "metadata:", "metadata:\n  namespace:", 1)
			line = line + "\n" + "  namespace: {{ .Release.Namespace }}"
			data[i] = line
		}
	}

	return data
}

func IsVersionExists(numaflowVersion string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/numaproj/numaflow/releases/tags/%s", numaflowVersion)
	_, err := api.DownloadFileDataWithRetry(url)
	if err != nil && strings.Contains(err.Error(), "404") {
		return false, err
	}

	return true, nil
}

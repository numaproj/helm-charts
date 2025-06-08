package internal

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/numaproj/helm-charts/upgrade/common"
	"helm.sh/helm/v3/pkg/chartutil"
)

// UpdateChartFile updates the Chart.yaml file with the new version and appVersion.
func UpdateChartFile(numaflowVersion string) {
	chartFilePath := common.BaseDir + chartutil.ChartfileName
	chart, err := chartutil.LoadChartfile(chartFilePath)
	if err != nil {
		log.Fatalf("Error loading chart file: %s\n", err)
	}

	numaflowVersion = strings.TrimPrefix(numaflowVersion, "v")
	if numaflowVersion == chart.AppVersion {
		log.Fatalln("Versions are identical. No update is necessary.")
	}
	numaflowParts := strings.Split(numaflowVersion, ".")
	appParts := strings.Split(chart.AppVersion, ".")
	currentParts := strings.Split(chart.Version, ".")

	majorUpdate := numaflowParts[0] != appParts[0]
	minorUpdate := !majorUpdate && numaflowParts[1] != appParts[1]
	patchUpdate := !majorUpdate && !minorUpdate && numaflowParts[2] != appParts[2]

	var newVersion string
	switch {
	case majorUpdate:
		currentVersion, err := strconv.Atoi(currentParts[0])
		if err != nil {
			log.Fatalf("Error parsing current version: %s\n", err)
		}
		newVersion = fmt.Sprintf("%d.0.0", currentVersion+1)
	case minorUpdate:
		currentVersion, err := strconv.Atoi(currentParts[1])
		if err != nil {
			log.Fatalf("Error parsing current version: %s\n", err)
		}
		newVersion = fmt.Sprintf("%s.%d.0", currentParts[0], currentVersion+1)
	case patchUpdate:
		currentVersion, err := strconv.Atoi(currentParts[2])
		if err != nil {
			log.Fatalf("Error parsing current version: %s\n", err)
		}
		newVersion = fmt.Sprintf("%s.%s.%d", currentParts[0], currentParts[1], currentVersion+1)
	default:
		log.Fatalf("No significant version change detected. No update to %s required.\n", numaflowVersion)
	}

	chart.Version = newVersion
	chart.AppVersion = numaflowVersion

	err = chartutil.SaveChartfile(chartFilePath, chart)
	if err != nil {
		log.Fatalf("Error saving chart file: %s\n", err)
	}

	log.Printf("Updated version: %s and appVersion: %s in charts/numaflow/Chart.yaml", newVersion, numaflowVersion)
}

// UpdateValuesFile updates the values.yaml file with the new Numaflow version.
func UpdateValuesFile(numaflowVersion string) {
	valuesFilePath := common.BaseDir + chartutil.ValuesfileName
	yamlContent, err := os.ReadFile(valuesFilePath)
	if err != nil {
		log.Fatalf("Error reading values file: %s\n", err)
	}

	lines := strings.Split(string(yamlContent), "\n")
	for i, line := range lines {
		if strings.Contains(line, "tag:") {
			// Find the index of "tag"
			index := strings.Index(line, "tag:")
			// Extract existing spaces before "tag"
			spaces := line[:index]
			// Replace the line, preserving the spaces
			lines[i] = fmt.Sprintf("%stag: %s", spaces, numaflowVersion)
			break
		}
	}

	updatedContent := strings.Join(lines, "\n")
	err = os.WriteFile(valuesFilePath, []byte(updatedContent), 0644)
	if err != nil {
		log.Fatalf("Error writing back to values file: %s\n", err)
	}

	log.Printf("Updated appVersion in charts/numaflow/values.yaml to %s", numaflowVersion)
}

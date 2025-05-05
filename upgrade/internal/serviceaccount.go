package internal

import (
	"fmt"
	"log"

	"github.com/numaproj/helm-charts/upgrade/common"
)

// UpdateServiceAccount updates the service account files in the local directory with the latest versions from the GitHub repository.
func UpdateServiceAccount(numaflowVersion string) {
	for fileName, url := range common.ServiceAccountFiles {
		localFilePath := generateServiceAccountFilePath(fileName, false)
		err := updateFiles(localFilePath, url, numaflowVersion, true)
		if err != nil {
			log.Printf("Error updating service account file: %s, err: %v\n", fileName, err)
			continue
		}
		log.Printf("Successfully updated service account file: %s ...\n", fileName)
	}
}

func generateServiceAccountFilePath(fileName string, namespaced bool) string {
	return fmt.Sprintf("%s%s%s", common.BaseDir, common.ServiceAccountBaseDir, fileName)
}

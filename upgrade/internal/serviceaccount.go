package internal

import (
	"fmt"

	"github.com/numaproj/helm-charts/upgrade/common"
)

func UpdateServiceAccount(numaflowVersion string) {
	for fileName, url := range common.ServiceAccountFiles {
		localFilePath := generateServiceAccountFilePath(fileName, false)
		err := updateFiles(localFilePath, url, numaflowVersion, true)
		if err != nil {
			fmt.Printf("Error updating service account file: %s, err: %v\n", fileName, err)
			continue
		}
		fmt.Printf("Successfully updated service account file: %s ...\n", fileName)
	}
}

func generateServiceAccountFilePath(fileName string, namespaced bool) string {
	return fmt.Sprintf("%s%s%s", common.BaseDir, common.ServiceAccountBaseDir, fileName)
}

package internal

import (
	"fmt"
	"github.com/numaproj/helm-charts/upgrade/common"
)

const crdsLocalPath = "crds/"

func UpdateCRDFiles(numaflowVersion string) {
	for fileName, path := range common.CRDFiles {
		githubFileUrl := path + fileName
		localFilePath := generateCRDsFilePath(fileName)
		if err := updateFiles(localFilePath, githubFileUrl, numaflowVersion, false); err != nil {
			println("Error updating CRD file:", fileName, "err:", err)
			continue
		}
		fmt.Println("Successfully updated CRD file:", fileName, "...")
	}
}

func generateCRDsFilePath(fileName string) string {
	return fmt.Sprintf("%s%s%s", common.BaseDir, crdsLocalPath, fileName)
}

package internal

import (
	"fmt"
	"log"

	"github.com/numaproj/helm-charts/upgrade/common"
)

const CRDSLocalPath = "crds/"

// UpdateCRDFiles updates the CRD files in the local directory with the latest versions from the GitHub repository.
func UpdateCRDFiles(numaflowVersion string) {
	for fileName, path := range common.CRDFiles {
		githubFileUrl := path + fileName
		localFilePath := generateCRDsFilePath(fileName)
		if err := updateFiles(localFilePath, githubFileUrl, numaflowVersion, false); err != nil {
			log.Println("Error updating CRD file:", fileName, "err:", err)
			continue
		}
		log.Println("Successfully updated CRD file:", fileName, "...")
	}
}

func generateCRDsFilePath(fileName string) string {
	return fmt.Sprintf("%s%s%s", common.BaseDir, CRDSLocalPath, fileName)
}

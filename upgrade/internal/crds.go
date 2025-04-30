package internal

import (
	"fmt"
	"github.com/numaproj/helm-charts/upgrade/common"
)

const crdsLocalPath = "charts/numaflow/crds/"

func UpdateCRDFiles(numaflowVersion string) {
	for fileName, path := range common.CRDFiles {
		fileUrl := path + fileName
		localFilePath := crdsLocalPath + fileName
		if err := updateFiles(localFilePath, fileUrl, numaflowVersion, false); err != nil {
			println("Error updating CRD file:", fileName, "err:", err)
			continue
		}
		fmt.Println("Successfully updated CRD file:", fileName, "...")
	}
}

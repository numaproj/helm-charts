package internal

import (
	"fmt"
	"log"

	"github.com/numaproj/helm-charts/upgrade/common"
)

// UpdateRBACFiles updates the RBAC files in the local directory with the latest versions from the GitHub repository.
func UpdateRBACFiles(numaflowVersion string) {
	// Update RBAC files for cluster-scoped resources
	for fileName, url := range common.RbacFilesForClusterScopedResources {
		localFilePath := generateRBACFilePath(fileName, false)
		if err := updateFiles(localFilePath, url, numaflowVersion, false); err != nil {
			log.Printf("Error updating cluster-scoped file: %s, err: %v\n", fileName, err)
			continue
		}
		fmt.Printf("Successfully updated cluster-scoped file: %s ...\n", fileName)
	}

	for fileName, url := range common.RbacFilesForNamespacedResources {
		localFilePath := generateRBACFilePath(fileName, true)
		if err := updateFiles(localFilePath, url, numaflowVersion, true); err != nil {
			log.Printf("Error updating namespaced file: %s, err: %v\n", fileName, err)
			continue
		}
		log.Printf("Successfully updated namespaced file: %s ...\n", fileName)
	}
}

func generateRBACFilePath(fileName string, namespaced bool) string {
	if namespaced {
		return fmt.Sprintf("%s%s%s", common.BaseDir, common.RBACNamespacedBaseDir, fileName)
	}
	return fmt.Sprintf("%s%s%s", common.BaseDir, common.RBACClusterScopedBaseDir, fileName)
}

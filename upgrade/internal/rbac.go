package internal

import (
	"fmt"

	"github.com/numaproj/helm-charts/upgrade/common"
)

func UpdateRBACFiles(numaflowVersion string) {
	// Update RBAC files for cluster-scoped resources
	for fileName, url := range common.RbacFilesForClusterScopedResources {
		localFilePath := generateRBACFilePath(fileName, false)
		if err := updateFiles(localFilePath, url, numaflowVersion, false); err != nil {
			fmt.Printf("Error updating cluster-scoped file: %s, err: %v\n", fileName, err)
			continue
		}
		fmt.Printf("Successfully updated cluster-scoped file: %s ...\n", fileName)
	}

	for fileName, url := range common.RbacFilesForNamespacedResources {
		localFilePath := generateRBACFilePath(fileName, true)
		if err := updateFiles(localFilePath, url, numaflowVersion, true); err != nil {
			fmt.Printf("Error updating namespaced file: %s, err: %v\n", fileName, err)
			continue
		}
		fmt.Printf("Successfully updated namespaced file: %s ...\n", fileName)
	}
}

func generateRBACFilePath(fileName string, namespaced bool) string {
	if namespaced {
		return fmt.Sprintf("%s%s%s", common.BaseDir, common.RBACNamespacedBaseDir, fileName)
	}
	return fmt.Sprintf("%s%s%s", common.BaseDir, common.RBACClusterScopedBaseDir, fileName)
}

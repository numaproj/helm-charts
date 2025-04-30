package common

import (
	"fmt"
	"os"
)

const (
	GithubBaseURL    = "https://raw.githubusercontent.com/numaproj/numaflow/"
	DefaultLabel     = "	{{- include \"numaflow.labels\" . | nindent 4 }}"
	ContentSeparator = "---"
)

var BaseDir string

func init() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	BaseDir = dir + "/charts/numaflow/templates/"
}

const (
	RBACClusterScopedBaseDir = "rbac/cluster-scoped/"
	RBACNamespacedBaseDir    = "rbac/namespaced/"
	ServiceAccountBaseDir    = "serviceaccounts/"
)

// Contains the mapping of RBAC files for cluster-scoped and namespaced resources
var (
	RbacFilesForClusterScopedResources = map[string]string{
		"numaflow-aggregate-to-admin-role.yaml":      "/config/cluster-install/rbac/controller-manager/numaflow-aggregate-to-admin.yaml",
		"numaflow-aggregate-to-edit-role.yaml":       "/config/cluster-install/rbac/controller-manager/numaflow-aggregate-to-edit.yaml",
		"numaflow-aggregate-to-view-role.yaml":       "/config/cluster-install/rbac/controller-manager/numaflow-aggregate-to-view.yaml",
		"numaflow-binding.yaml":                      "/config/cluster-install/rbac/controller-manager/numaflow-binding.yaml",
		"numaflow-cluster-role.yaml":                 "/config/cluster-install/rbac/controller-manager/numaflow-cluster-role.yaml",
		"numaflow-server-binding.yaml":               "/config/cluster-install/rbac/numaflow-server/numaflow-server-binding.yaml",
		"numaflow-server-cluster-role.yaml":          "/config/cluster-install/rbac/numaflow-server/numaflow-server-cluster-role.yaml",
		"numaflow-webhook-cluster-role.yaml":         "/config/extensions/webhook/rbac/numaflow-webhook-cluster-role.yaml",
		"numaflow-webhook-cluster-role-binding.yaml": "/config/extensions/webhook/rbac/numaflow-webhook-cluster-role-binding.yaml",
	}

	RbacFilesForNamespacedResources = map[string]string{
		"numaflow-dex-server-rolebinding.yaml": "/config/base/dex/numaflow-dex-server-rolebinding.yaml",
		"numaflow-dex-server-role.yaml":        "/config/base/dex/numaflow-dex-server-role.yaml",
		"numaflow-role.yaml":                   "/config/namespace-install/rbac/controller-manager/numaflow-role.yaml",
		"numaflow-binding.yaml":                "/config/namespace-install/rbac/controller-manager/numaflow-binding.yaml",
		"numaflow-server-binding.yaml":         "/config/namespace-install/rbac/numaflow-server/numaflow-server-binding.yaml",
		"numaflow-server-role.yaml":            "/config/namespace-install/rbac/numaflow-server/numaflow-server-role.yaml",
		"numaflow-server-secrets-binding.yaml": "/config/namespace-install/rbac/numaflow-server/numaflow-server-secrets-binding.yaml",
		"numaflow-server-secrets-role.yaml":    "/config/namespace-install/rbac/numaflow-server/numaflow-server-secrets-role.yaml",
	}
)

// ServiceAccountFiles Contains the mapping of service account files
var (
	ServiceAccountFiles = map[string]string{
		"numaflow-dex-server.yaml": "/config/base/dex/numaflow-dex-server-sa.yaml",
		"numaflow-sa.yaml":         "/config/base/controller-manager/numaflow-sa.yaml",
		"numaflow-server-sa.yaml":  "/config/base/numaflow-server/numaflow-server-sa.yaml",
		"numaflow-webhook-sa.yaml": "/config/extensions/webhook/numaflow-webhook-sa.yaml",
	}
)

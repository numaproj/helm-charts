package common

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	GithubBaseURL = "https://raw.githubusercontent.com/numaproj/numaflow/"
	DefaultLabel  = "	{{- include \"numaflow.labels\" . | nindent 4 }}"
)

// BaseDir is the absolute path to charts/numaflow/, with a trailing slash so
// callers can string-concatenate sub-paths the same way they did before.
var BaseDir string

// init locates charts/numaflow/ by walking up from CWD. This makes the tool
// runnable from the repo root, from upgrade/, or from any deeper directory —
// previously it only worked when invoked from upgrade/.
func init() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error resolving working directory:", err)
		return
	}
	found, ok := findChartsDir(cwd)
	if !ok {
		// Fall back to the legacy behaviour so existing flows keep working
		// even if the walker doesn't find a chart dir (e.g. tooling invoked
		// from an entirely unrelated location).
		BaseDir = cwd + "/../charts/numaflow/"
		return
	}
	BaseDir = found + string(filepath.Separator)
}

// findChartsDir walks up from start looking for `charts/numaflow/`. Returns
// the absolute path to that directory (no trailing separator) and true on
// success.
func findChartsDir(start string) (string, bool) {
	dir := start
	for {
		candidate := filepath.Join(dir, "charts", "numaflow")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

const (
	RBACClusterScopedBaseDir = "templates/rbac/cluster-scoped/"
	RBACNamespacedBaseDir    = "templates/rbac/namespaced/"
	ServiceAccountBaseDir    = "templates/serviceaccounts/"
)

var (
	// RbacFilesForClusterScopedResources contains the mapping of RBAC files for cluster-scoped resources
	// Here key is the file name and value is the file path in the GitHub repository
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

	// RbacFilesForNamespacedResources contains the mapping of RBAC files for namespaced resources
	// Here key is the file name and value is the path in the GitHub repository
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

	// ServiceAccountFiles contains the mapping of service account files to their respective paths
	// Here key is the file name and value is the path in the GitHub repository
	ServiceAccountFiles = map[string]string{
		"numaflow-dex-server.yaml": "/config/base/dex/numaflow-dex-server-sa.yaml",
		"numaflow-sa.yaml":         "/config/base/controller-manager/numaflow-sa.yaml",
		"numaflow-server-sa.yaml":  "/config/base/numaflow-server/numaflow-server-sa.yaml",
		"numaflow-webhook-sa.yaml": "/config/extensions/webhook/numaflow-webhook-sa.yaml",
	}

	// CRDFiles contains the mapping of CRD files to their respective paths
	// Here key is the file name and value is the path in the GitHub repository
	CRDFiles = map[string]string{
		"numaflow.numaproj.io_interstepbufferservices.yaml": "/config/base/crds/full/",
		"numaflow.numaproj.io_pipelines.yaml":               "/config/base/crds/full/",
		"numaflow.numaproj.io_vertices.yaml":                "/config/base/crds/full/",
		"numaflow.numaproj.io_monovertices.yaml":            "/config/base/crds/full/",
		"numaflow.numaproj.io_servingpipelines.yaml":        "/config/base/crds/full/",
	}
)

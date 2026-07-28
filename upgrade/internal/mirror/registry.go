// Package mirror automates the maintenance of the 15 chart files that mirror
// upstream numaflow resources but cannot be auto-overwritten because they
// carry Helm templating (configmaps, deployments, secrets, services).
//
// The package exposes two layers (named L1/L2 in the design doc):
//
//   - Layer 1 (upstream.go) — upstream-vs-upstream diff: for each registered
//     mirror, fetch the upstream blob at both vOld and vNew. If they are
//     byte-identical there is no work to do.
//
//   - Layer 2 (threeway.go) — git 3-way merge: when the upstream blob has
//     changed, apply that change onto the chart's templated copy using
//     `git merge-file --diff3` with the chart's Helm expressions first
//     normalised to deterministic sentinels (tokenize.go) so that templating
//     does not cause spurious conflicts.
package mirror

// MirroredFile describes a chart file that is hand-mirrored from upstream
// numaflow but carries Helm templating, so the existing upgrade tool cannot
// safely overwrite it.
type MirroredFile struct {
	// LocalPath is the path to the chart file, relative to common.BaseDir
	// (i.e. relative to charts/numaflow/).
	LocalPath string

	// UpstreamPath is the path inside the numaproj/numaflow repository,
	// appended to common.GithubBaseURL + <version>.
	UpstreamPath string
}

// MirroredFiles is the canonical registry of the chart files that mirror
// upstream numaflow resources. Entries are ordered by category
// (configmaps, deployments, secrets, services) and then alphabetically
// within each category to make additions easy to review.
//
// To add a new mirrored file, append an entry here. The `sync` subcommand
// picks it up automatically — no other code changes are required.
var MirroredFiles = []MirroredFile{
	// ---- configmaps ----
	{
		LocalPath:    "templates/configmaps/numaflow-cmd-params-config.yaml",
		UpstreamPath: "/config/base/shared-config/numaflow-cmd-params-config.yaml",
	},
	{
		LocalPath:    "templates/configmaps/numaflow-controller-config.yaml",
		UpstreamPath: "/config/base/controller-manager/numaflow-controller-config.yaml",
	},
	{
		LocalPath:    "templates/configmaps/numaflow-dex-server-config.yaml",
		UpstreamPath: "/config/base/dex/numaflow-dex-server-configmap.yaml",
	},
	{
		LocalPath:    "templates/configmaps/numaflow-server-local-user-config.yaml",
		UpstreamPath: "/config/base/numaflow-server/numaflow-server-local-user-config.yaml",
	},
	{
		LocalPath:    "templates/configmaps/numaflow-server-metrics-proxy-config.yaml",
		UpstreamPath: "/config/base/numaflow-server/numaflow-server-metrics-proxy-config.yaml",
	},
	{
		LocalPath:    "templates/configmaps/numaflow-server-rbac-config.yaml",
		UpstreamPath: "/config/base/numaflow-server/numaflow-server-rbac-config.yaml",
	},

	// ---- deployments ----
	{
		LocalPath:    "templates/deployments/numaflow-controller.yaml",
		UpstreamPath: "/config/base/controller-manager/controller-manager-deployment.yaml",
	},
	{
		LocalPath:    "templates/deployments/numaflow-dex-server.yaml",
		UpstreamPath: "/config/base/dex/numaflow-dex-server-deployment.yaml",
	},
	{
		LocalPath:    "templates/deployments/numaflow-server.yaml",
		UpstreamPath: "/config/base/numaflow-server/numaflow-server-deployment.yaml",
	},
	{
		LocalPath:    "templates/deployments/numaflow-webhook.yaml",
		UpstreamPath: "/config/extensions/webhook/numaflow-webhook-deployment.yaml",
	},

	// ---- secrets ----
	{
		LocalPath:    "templates/secrets/numaflow-dex-secrets.yaml",
		UpstreamPath: "/config/base/dex/numaflow-dex-secrets.yaml",
	},
	{
		LocalPath:    "templates/secrets/numaflow-server-secrets.yaml",
		UpstreamPath: "/config/base/numaflow-server/numaflow-server-secrets.yaml",
	},

	// ---- services ----
	{
		LocalPath:    "templates/services/numaflow-dex-server.yaml",
		UpstreamPath: "/config/base/dex/numaflow-dex-server-service.yaml",
	},
	{
		LocalPath:    "templates/services/numaflow-server.yaml",
		UpstreamPath: "/config/base/numaflow-server/numaflow-server-service.yaml",
	},
	{
		// NB: MAINTAINERS.md historically pointed this at numaflow-webhook-sa.yaml,
		// which is a ServiceAccount, not a Service. The correct upstream is the
		// webhook service definition.
		LocalPath:    "templates/services/numaflow-webhook.yaml",
		UpstreamPath: "/config/extensions/webhook/numaflow-webhook-service.yaml",
	},
}

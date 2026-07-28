# Helm Chart for NUMAFLOW Project

This repository contains the Helm charts for deploying Numaflow. As the Numaflow project evolves with new versions, updates to this Helm chart may be necessary to accommodate new features, improvements, or breaking changes.

> **Note:** Releasing a new helm chart is partially automated. `make upgrade-charts` refreshes Chart.yaml, values.yaml, CRDs, RBAC, and ServiceAccounts; `make sync-mirrored-files` handles the configmaps, deployments, secrets, and services that previously had to be hand-diffed.

## Updating Helm Chart for New Numaflow Versions

**Step 1: refresh auto-managed files**

```
NUMAFLOW_VERSION=vx.y.z make upgrade-charts
```

This bumps `version`/`appVersion` in `Chart.yaml`, updates the image `tag` in `values.yaml`, and pulls fresh CRDs, RBAC, and ServiceAccount definitions from upstream.

**Step 2: sync the mirrored chart files**

```
NUMAFLOW_VERSION=vx.y.z make sync-mirrored-files
```

This compares each manually-mirrored file (the 15 configmaps, deployments, secrets, and services tracked in [`upgrade/internal/mirror/registry.go`](upgrade/internal/mirror/registry.go)) against `numaproj/numaflow` at the previous and new versions, applies upstream changes onto the templated chart copies via a 3-way merge, and writes the result.

By default the run is dry: every clean merge is written to `upgrade/.merge-rejects/<file>.merged`, and any conflict is written to `upgrade/.merge-rejects/<file>.conflict`. Review the output, then re-run with `--apply` to write the clean merges back to the chart:

```
NUMAFLOW_VERSION=vx.y.z make sync-mirrored-files SYNC_FLAGS=--apply
```

Conflicts (`<<<<<<< / ||||||| / ======= / >>>>>>>` markers) require human resolution: open the `*.conflict` file, decide whether to keep the chart's templated value or the upstream change, copy the resolved content into the chart file, and delete the marker file.

Useful flags (pass via `SYNC_FLAGS`):
- `--from-version=vx.y.z` — override the baseline version (defaults to the `appVersion` in the committed `HEAD` `Chart.yaml`, i.e. the previous release, so it is unaffected by the `upgrade-charts` bump in Step 1). Useful when catching up several releases at once.
- `--only=file1.yaml,file2.yaml` — restrict the run to a subset of files (basenames).

**Step 3: verify locally**

```
helm lint charts/numaflow
helm template charts/numaflow --namespace numaflow-system | head
helm install numaflow charts/numaflow --namespace numaflow-system --create-namespace
```

**Step 4: open a PR**

Create a Pull Request against `main`. Once it is merged, CI publishes the new helm chart release at <https://github.com/numaproj/helm-charts/releases>.

**Step 5: install from the published chart**

Follow [these](./charts/numaflow/README.md) steps to install and verify the helm chart in your cluster.

Happy helming!!!

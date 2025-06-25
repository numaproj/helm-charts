# Helm Chart for NUMAFLOW Project

This repository contains the Helm charts for deploying Numaflow. As the Numaflow project evolves with new versions, updates to this Helm chart may be necessary to accommodate new features, improvements, or breaking changes.

> **Note:** This repo contains the partial automation which helps in updating the helm chart for new numaflow versions.
Like updating the CRDs, RBAC, ServiceAccount. The rest of the files (Deployments, Secrets, Services and Configmaps) need to be updated manually by comparing it with the current version of numaflow.

## Updating Helm Chart for New Numaflow Versions

**Step 1:**
- Update Numaflow CRDs, RBACs and ServiceAccounts. This will also upgrade the version and AppVersion in the `Chart.yaml` file.
```
NUMAFLOW_VERSION=v1.4.4 make upgrade-charts
```

**Step 3:**
- Update these file changes accordingly by comparing it with the current Numaflow Version.
  - [configmaps](charts/numaflow/templates/configmaps)
  - [deployments](charts/numaflow/templates/deployments)
  - [secrets](./charts/numaflow/templates/secrets)
  - [services](./charts/numaflow/templates/services)

**Example:** The transition from version `v1.4.0` to `v1.4.4` includes updates to the Configmap, which are detailed in [PR #28](https://github.com/numaproj/helm-charts/pull/28/files). This PR reflects the changes made by comparing the existing [Configmap file](charts/numaflow/templates/configmaps/numaflow-server-metrics-proxy-config.yaml) for `numaflow-server-metrics-proxy-config` with the [Configmap for version v1.4.4](https://github.com/numaproj/numaflow/blob/v1.4.4/config/base/numaflow-server/numaflow-server-metrics-proxy-config.yaml).
Similar for other files, you can compare the changes in the respective files in the `numaflow` repo.

**Step 4:**
- Verify the changes by running below helm command in local k8s cluster
```
helm install numaflow charts/numaflow --namespace numaflow-system --create-namespace
```

**Step 5:**
- Create a Pull Request against `main` branch and wait for it to get merged, Once it's merged, `CI` will publish the new helm chart release [here](https://github.com/numaproj/helm-charts/releases).

**Step 6:**
- Follow [these](./charts/numaflow/README.md) steps to install and verify the helm chart in your cluster.

Happy helming!!!

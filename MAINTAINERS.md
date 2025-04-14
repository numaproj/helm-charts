# Helm Chart for NUMAFLOW Project

This repository contains the Helm charts for deploying Numaflow. As the Numaflow project evolves with new versions, updates to this Helm chart may be necessary to accommodate new features, improvements, or breaking changes.

## Updating Helm Chart for New Numaflow Versions

**Step 1:**
- Update Numaflow CRDs (This will pull the changes from [here](https://github.com/numaproj/numaflow/tree/main/config/base/crds/full) based on the version specified in the `NUMAFLOW_VERSION` variable)
```
NUMAFLOW_VERSION=v1.4.4 make update-crds
```

**Step 2:**
- Update the [Chart.yaml](./charts/numaflow/Chart.yaml) file with the new `appVersion`(Numaflow version) and increment the `version` accordingly.

**Step 3:**
- Update the [values.yaml](./charts/numaflow/values.yaml) file with the new `image tag`. The image tag should be the same as the version specified in the `NUMAFLOW_VERSION` variable.

**Step 4:**
- Update these file changes accordingly by comparing it with the current Numaflow Version.
  - [configmap.yaml](./charts/numaflow/templates/configmap.yaml)
  - [deployment.yaml](./charts/numaflow/templates/deployment.yaml)
  - [role.yaml](./charts/numaflow/templates/role.yaml)
  - [rolebinding.yaml](./charts/numaflow/templates/rolebinding.yaml)
  - [secret.yaml](./charts/numaflow/templates/secret.yaml)
  - [service.yaml](./charts/numaflow/templates/service.yaml)
  - [serviceaccount.yaml](./charts/numaflow/templates/serviceaccount.yaml)

**Step 5:**
- Verify the changes by running below helm command in local k8s cluster
```
helm install numaflow charts/numaflow --namespace numaflow-system --create-namespace
```

**Step 6:**
- Create a Pull Request against `main` branch and wait for it to get merged, Once it's merged, `CI` will publish the new helm chart release [here](https://github.com/numaproj/helm-charts/releases).

**Step 7:**
- Follow [these](./charts/numaflow/README.md) steps to install and verify the helm chart in your cluster.
- Happy helming!!!

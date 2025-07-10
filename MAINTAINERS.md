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
- Update these file changes accordingly by comparing it with the current Numaflow Version. (Note: All below examples are based on `v1.5.1` version of Numaflow, So change the version accordingly while comparing)
  - Update the [configmaps](charts/numaflow/templates/configmaps)
    - File [numaflow-cmd-params-config.yaml](charts/numaflow/templates/configmaps/numaflow-cmd-params-config.yaml) should be updated with numaflow [numaflow-cmd-params-config.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/base/shared-config/numaflow-cmd-params-config.yaml)
    - File [numaflow-server-metrics-proxy-config.yaml](charts/numaflow/templates/configmaps/numaflow-server-metrics-proxy-config.yaml) should be updated with [numaflow-server-metrics-proxy-config.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/base/numaflow-server/numaflow-server-metrics-proxy-config.yaml)
    - File [numaflow-controller-config.yaml](charts/numaflow/templates/configmaps/numaflow-controller-config.yaml) should be updated with [numaflow-controller-config.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/base/controller-manager/numaflow-controller-config.yaml)
    - File [numaflow-dex-server-config.yaml](charts/numaflow/templates/configmaps/numaflow-dex-server-config.yaml) should be updated with [numaflow-dex-server-config.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/base/dex/numaflow-dex-server-configmap.yaml)
    - File [numaflow-server-local-uer-config.yaml](charts/numaflow/templates/configmaps/numaflow-server-local-user-config.yaml) should be updated with [numaflow-server-local-user-config.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/base/numaflow-server/numaflow-server-local-user-config.yaml)
    - File [numaflow-server-rbac-config.yaml](charts/numaflow/templates/configmaps/numaflow-server-rbac-config.yaml) should be updated with [numaflow-server-rbac-config.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/base/numaflow-server/numaflow-server-rbac-config.yaml)
  - Update the [deployments](charts/numaflow/templates/deployments)
    - File [numaflow-controller.yaml](charts/numaflow/templates/deployments/numaflow-controller.yaml) should be updated with [numaflow-controller.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/base/controller-manager/controller-manager-deployment.yaml)
    - File [numaflow-server.yaml](charts/numaflow/templates/deployments/numaflow-server.yaml) should be updated with [numaflow-server.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/base/numaflow-server/numaflow-server-deployment.yaml)
    - FIle [numaflow-dex-server.yaml](charts/numaflow/templates/deployments/numaflow-dex-server.yaml) should be updated with [numaflow-dex-server.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/base/dex/numaflow-dex-server-deployment.yaml)
    - File [numaflow-webhook.yaml](charts/numaflow/templates/deployments/numaflow-webhook.yaml) should be updated with [numaflow-webhook.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/extensions/webhook/numaflow-webhook-deployment.yaml)w
  - Update the [secrets](./charts/numaflow/templates/secrets)
    - File [numaflow-dex-secrets.yaml](charts/numaflow/templates/secrets/numaflow-dex-secrets.yaml) should be updated with [numaflow-dex-secret.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/base/dex/numaflow-dex-secrets.yaml)
    - File [numaflow-server-secrets.yaml](charts/numaflow/templates/secrets/numaflow-server-secrets.yaml) should be updated with [numaflow-server-secrets.yaml](s
  - Update the [services](./charts/numaflow/templates/services)
    - File [numaflow-dex-server.yaml](charts/numaflow/templates/services/numaflow-dex-server.yaml) should be updated with [numaflow-dex-server.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/base/dex/numaflow-dex-server-service.yaml)
    - File [numaflow-server.yaml](charts/numaflow/templates/services/numaflow-server.yaml) should be updated with [numaflow-server.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/base/numaflow-server/numaflow-server-service.yaml)
    - File [numaflow-webhook.yaml](charts/numaflow/templates/services/numaflow-webhook.yaml) should be updated with [numaflow-webhook.yaml](https://github.com/numaproj/numaflow/blob/v1.5.1/config/extensions/webhook/numaflow-webhook-sa.yaml)

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

# helm-charts
Helm charts for the Numaproj ecosystem projects.

## Usages
- Please refer to the [README](./charts/numaflow/README.md) file in the `charts/numaflow` directory for detailed instructions on how to install and use the Numaflow Helm chart.

## Versions

The released chart versions and related Numaflow app versions are:

```text
NAME                    CHART VERSION   APP VERSION     DESCRIPTION
numaproj/numaflow       0.2.1           1.5.1           A Helm chart for installing Numaflow in Kubernetes
numaproj/numaflow       0.2.0           1.5.0           A Helm chart for installing Numaflow in Kubernetes
numaproj/numaflow       0.0.7           1.4.4           A Helm chart for installing Numaflow in Kubernetes
numaproj/numaflow       0.0.6           1.4.0           A Helm chart for installing Numaflow in Kubernetes
numaproj/numaflow       0.0.5           1.3.3           A Helm chart for installing Numaflow in Kubernetes
numaproj/numaflow       0.0.4                           A Helm chart for installing Numaflow in Kubernetes
numaproj/numaflow       0.0.3                           A Helm chart for installing Numaflow in Kubernetes
numaproj/numaflow       0.0.2                           A Helm chart for installing Numaflow in Kubernetes
numaproj/numaflow       0.0.1                           A Helm chart for installing Numaflow in Kubernetes
```

For an up-to-date list, run:

```bash
helm repo add numaproj https://numaproj.io/helm-charts
helm repo update numaproj
helm search repo numaproj/numaflow --versions
```

## Contributing
- Please refer to the [MAINTAINERS](./MAINTAINERS.md) document for guidelines on how to contribute to this repository.

# Numaflow helm chart
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A Helm chart for installing Numaflow in Kubernetes

## Usage

### PreRequisites

- Install [helm3](https://helm.sh/docs/intro/install/)

### Installation Steps

The following steps will help you install numaflow via helm.

#### Step-1: Add the numaflow helm repository

```bash
helm repo add numaflow https://numaproj.io/helm-charts
```
```bash
helm repo ls

Output:

NAME         URL
numaflow     https://numaproj.io/helm-charts
```

#### Step-2: Install the Numaflow

```bash
helm install numaflow numaflow/numaflow --namespace numaflow-system --create-namespace
```

> **_NOTE:_**  By default numaflow will be installed in Cluster Scope, for namespace scope installation run below command.
> ```bash
> helm install numaflow numaflow/numaflow --namespace numaflow-system --set server.configs.namespacedScope=true --create-namespace
> ```

Output:

```bash
NAME: numaflow
LAST DEPLOYED: Thu Jan 18 19:51:50 2024
NAMESPACE: numaflow-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
1. Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace numaflow-system -l "app.kubernetes.io/name=numaflow-server" -o jsonpath="{.items[0].metadata.name}")
  export CONTAINER_PORT=$(kubectl get pod --namespace numaflow-system $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
  kubectl --namespace numaflow-system port-forward $POD_NAME $CONTAINER_PORT
  echo "Visit http://127.0.0.1:$CONTAINER_PORT to use your application"
```

you have successfully installed Numaflow!!!

### Additional Steps (Verification)
You can run the following commands if you wish to verify if all desired components are installed successfully.

- Check if the numaflow components are running successfully
```bash
kubectl get po -n numaflow-system

Output:

NAME                                   READY   STATUS    RESTARTS   AGE
numaflow-controller-6f764f7958-mfvq7   1/1     Running   0          75s
numaflow-dex-server-c785467c7-c6hqj    1/1     Running   0          75s
numaflow-server-c58445b7b-lxjw4        1/1     Running   0          75s
numaflow-webhook-74477447cc-b255v      1/1     Running   0          75s
```

### Run a simple pipeline

Follow the steps [here](https://numaflow.numaproj.io/quick-start/#creating-a-simple-pipeline)

### Uninstalling the Numaflow

```bash
helm uninstall numaflow --namespace numaflow-system
```

### Others values can be overridden using below configuration.
| Key | Type | Default | Description |
|-----|------|---------|-------------|
| controller.replicaCount | int | `1` | The number of controller replicas to run. |
| controller.resources.limits.cpu | string | `"500m"` | The CPU limits for controller. |
| controller.resources.limits.memory | string | `"1024Mi"` | The memory limits for controller. |
| controller.resources.requests.cpu | string | `"100m"` | The CPU requests for controller. |
| controller.resources.requests.memory | string | `"200Mi"` | The memory requests for controller. |
| dexServer.image.pullPolicy | string | `"Always"` | Image Pull policy of dex server for authentication. |
| dexServer.image.repository | string | `"dexidp/dex"` | Image of dex server for authentication. |
| dexServer.image.tag | string | `"v2.37.0"` | Tag of dex server for authentication. |
| dexServer.replicaCount | int | `1` |  |
| dexServer.secret.data.GITHUB_CLIENT_ID | string | `""` | GitHub client ID for authentication. |
| dexServer.secret.data.GITHUB_CLIENT_SECRET | string | `""` | GitHub client secret for authentication. |
| numaflow.image.pullPolicy | string | `"Always"` | Image Pull policy of numaflow server. |
| numaflow.image.repository | string | `"quay.io/numaproj/numaflow"` | Image of numaflow server. |
| numaflow.image.tag | string | `"v1.1.1"` | Tag of numaflow server. |
| server.configs.address | string | `"https://localhost:8443"` | The external address of the Numaflow server. This is needed when using Dex for authentication. |
| server.configs.authDisable | bool | `false` | Whether to disable authentication and authorization for the UX server, defaults to false. |
| server.configs.baseHref | string | `"/"` | Base href for Numaflow UX server, defaults to '/'. |
| server.configs.dexServer | string | `"http://numaflow-dex-server:5556/dex"` | The address of the Dex server for authentication. |
| server.configs.insecure | bool | `false` | Whether to disable TLS for UX server. |
| server.configs.leaderElection | bool | `false` | Whether to disable leader election for the controller, defaults to false |
| server.configs.managedNamespace | string | `"numaflow-system"` | The namespace that the controller and the UX server watch when "namespaced" is true. |
| server.configs.namespacedScope | bool | `false` | Whether to run the controller and the UX server in namespaced scope, defaults to false. |
| server.configs.port | int | `8443` | Port to listen on for UX server, defaults to 8443 or 8080 if insecure is set. |
| server.replicaCount | int | `1` | The number of numaflow-server replicas to run. |
| server.resources.limits.cpu | string | `"500m"` | The CPU limits for numaflow-server. |
| server.resources.limits.memory | string | `"1024Mi"` | The memory limits for numaflow-server. |
| server.resources.requests.cpu | string | `"100m"` | The CPU requests for numaflow-server. |
| server.resources.requests.memory | string | `"200Mi"` | The memory requests for numaflow-server. |
| server.service.port | int | `8443` | The port of the numaflow server. |
| server.service.type | string | `"ClusterIP"` | The type of service for the numaflow server. |

----------------------------------------------

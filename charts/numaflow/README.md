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
| Key                                        | Type   | Default                                 | Description                                                                                    |
|--------------------------------------------|--------|-----------------------------------------|------------------------------------------------------------------------------------------------|
| controller.replicaCount                    | int    | `1`                                     | The number of controller replicas to run.                                                      |
| controller.resources.limits.cpu            | string | `"500m"`                                | The CPU limits for controller.                                                                 |
| controller.resources.limits.memory         | string | `"1024Mi"`                              | The memory limits for controller.                                                              |
| controller.resources.requests.cpu          | string | `"100m"`                                | The CPU requests for controller.                                                               |
| controller.resources.requests.memory       | string | `"200Mi"`                               | The memory requests for controller.                                                            |
| dexServer.image.pullPolicy                 | string | `"Always"`                              | Image Pull policy of dex server for authentication.                                            |
| dexServer.image.repository                 | string | `"dexidp/dex"`                          | Image of dex server for authentication.                                                        |
| dexServer.image.tag                        | string | `"v2.37.0"`                             | Tag of dex server for authentication.                                                          |
| dexServer.replicaCount                     | int    | `1`                                     |                                                                                                |
| dexServer.secret.data.GITHUB_CLIENT_ID     | string | `""`                                    | GitHub client ID for authentication.                                                           |
| dexServer.secret.data.GITHUB_CLIENT_SECRET | string | `""`                                    | GitHub client secret for authentication.                                                       |
| numaflow.image.pullPolicy                  | string | `"Always"`                              | Image Pull policy of numaflow server.                                                          |
| numaflow.image.repository                  | string | `"quay.io/numaproj/numaflow"`           | Image of numaflow server.                                                                      |
| numaflow.image.tag                         | string | `"v1.1.1"`                              | Tag of numaflow server.                                                                        |
| server.configs.address                     | string | `"https://localhost:8443"`              | The external address of the Numaflow server. This is needed when using Dex for authentication. |
| server.configs.authDisable                 | bool   | `false`                                 | Whether to disable authentication and authorization for the UX server, defaults to false.      |
| server.configs.baseHref                    | string | `"/"`                                   | Base href for Numaflow UX server, defaults to '/'.                                             |
| server.configs.dexServer                   | string | `"http://numaflow-dex-server:5556/dex"` | The address of the Dex server for authentication.                                              |
| server.configs.insecure                    | bool   | `false`                                 | Whether to disable TLS for UX server.                                                          |
| server.configs.leaderElection              | bool   | `false`                                 | Whether to disable leader election for the controller, defaults to false                       |
| server.configs.managedNamespace            | string | `"numaflow-system"`                     | The namespace that the controller and the UX server watch when "namespaced" is true.           |
| server.configs.namespacedScope             | bool   | `false`                                 | Whether to run the controller and the UX server in namespaced scope, defaults to false.        |
| server.configs.port                        | int    | `8443`                                  | Port to liAof service for the numaflow server.                                                 |

----------------------------------------------

## How to contribute

- Step 1: Clone [helm-chart](https://github.com/numaproj/helm-charts) repository in local and checkout a new branch from `main` branch.
- Step 2: Make the required changes in `charts/numaflow` accordingly.
  - Note: Make sure to update the version in `chart/numaflow/Chart.yaml` according to [sem versioning](https://semver.org/)
- Step 3: Get the changes merged in `main` branch
- Step 4: Create a helm package using `helm package charts/numaflow` from `main` branch on latest changes, it should create a package like `numaflow-x.x.x.tgz`.
- Step 5: Checkout to branch `gh-pages` and move `numaflow-x.x.x.tgz` in dir `numaflow`.
- Step 6: Run command `helm repo index numaflow --merge index.yaml --url https://numaproj.io/helm-charts`, it will generate a new `numaflow/index.yaml`
- Step 7: Run command `mv -f numaflow/index.yaml` to update the existing `index.yaml`
- Step 8: Fix the path of numaflow chart url from `https://numaproj.io/helm-charts/numaflow-x.x.x.tgz` to `https://numaproj.io/helm-charts/numaflow/numaflow-x.x.x.tgz` in `index.yaml`
- Step 9: Commit and raise the PR to get the changes merged.
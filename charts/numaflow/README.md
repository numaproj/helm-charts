# Numaflow Helm Chart
A Helm chart for installing Numaflow in Kubernetes

## Values
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

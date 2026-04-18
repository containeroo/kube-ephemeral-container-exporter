# Helm Chart Values for kube-ephemeral-container-exporter

This chart renders the same default resources as `deploy/kubernetes/manifests`, including the namespace, monitoring resources, and Cilium policies.

## Important defaults

| Key | Description | Default |
| --- | --- | --- |
| `namespace.create` | Create the release namespace resource. | `true` |
| `image.repository` | Exporter image repository. | `ghcr.io/containeroo/kube-ephemeral-container-exporter` |
| `image.tag` | Exporter image tag. | `latest` |
| `image.pullPolicy` | Image pull policy. | `Always` |
| `replicas` | Deployment replica count. | `1` |
| `watchNamespaces` | Namespaces to watch. Empty means cluster-wide. | `[]` |
| `leaderElection.enabled` | Enable controller leader election. | `true` |

## Deployment and pod values

| Key | Description | Default |
| --- | --- | --- |
| `resources` | Container resource requests and limits. | See `values.yaml` |
| `podAnnotations` | Pod annotations applied to the deployment. | includes `kubectl.kubernetes.io/default-container` |
| `podLabels` | Extra Pod labels. | `{}` |
| `podSecurityContext` | Pod security context. | `runAsNonRoot: true` |
| `securityContext` | Container security context. | drop all caps, no privilege escalation |
| `startupProbe.enabled` | Enable startup probe. | `false` |
| `livenessProbe.enabled` | Enable liveness probe. | `true` |
| `readinessProbe.enabled` | Enable readiness probe. | `true` |
| `probes.address` | Override the health probe bind address. | `""` |
| `server.enableHTTP2` | Enable HTTP/2 for the metrics and probe servers. | `false` |
| `env` | Extra environment variables for the exporter container. | `[]` |
| `extraArgs` | Additional CLI arguments appended to the command. | `[]` |

## Monitoring values

| Key | Description | Default |
| --- | --- | --- |
| `metrics.enabled` | Enable the metrics endpoint. | `true` |
| `metrics.address` | Override the metrics bind address. | `""` |
| `metrics.secure` | Serve metrics over HTTPS. | `false` |
| `metrics.service.enabled` | Create the metrics Service. | `true` |
| `metrics.serviceMonitor.enabled` | Create the ServiceMonitor. | `true` |
| `metrics.serviceMonitor.jobLabel` | Service label used as the Prometheus `job`. | `app.kubernetes.io/name` |
| `metrics.prometheusRule.enabled` | Create the PrometheusRule. | `true` |
| `metrics.prometheusRule.namespace` | Namespace for the PrometheusRule. Empty uses the release namespace. | `""` |
| `metrics.prometheusRule.exporterDown.enabled` | Alert when no healthy exporter target is scraped. | `true` |
| `metrics.prometheusRule.longRunning.enabled` | Alert when an ephemeral container keeps running too long. | `true` |
| `metrics.prometheusRule.longRunning.for` | Duration before firing the long-running alert. | `1h` |

## RBAC and networking values

| Key | Description | Default |
| --- | --- | --- |
| `clusterRole.create` | Create the ClusterRole and ClusterRoleBinding. | `true` |
| `clusterRole.name` | Existing ClusterRole name to bind to. | `""` |
| `serviceAccount.create` | Create the ServiceAccount. | `true` |
| `serviceAccount.annotations` | ServiceAccount annotations. | `{}` |
| `ciliumNetworkPolicy.enabled` | Render the CiliumNetworkPolicy resources. | `true` |
| `ciliumNetworkPolicy.prometheusNamespace` | Namespace of the Prometheus pods allowed to scrape. | `observability` |

## Misc values

| Key | Description | Default |
| --- | --- | --- |
| `imagePullSecrets` | Image pull secrets for private registries. | `[]` |
| `nameOverride` | Short name override. | `""` |
| `fullnameOverride` | Full release name override. | `""` |
| `extraObjects` | Extra Kubernetes objects rendered with the chart. | `[]` |

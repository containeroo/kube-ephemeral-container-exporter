# Helm Chart Values for kube-ephemeral-container-exporter

This chart deploys `kube-ephemeral-container-exporter`, which watches Pods and exposes Prometheus metrics for attached ephemeral containers.

## Core Values

| Key                      | Description                                             | Default                                                 |
| ------------------------ | ------------------------------------------------------- | ------------------------------------------------------- |
| `image.repository`       | Exporter image repository.                              | `ghcr.io/containeroo/kube-ephemeral-container-exporter` |
| `image.tag`              | Image tag override. Uses chart `appVersion` when empty. | `""`                                                    |
| `image.pullPolicy`       | Image pull policy.                                      | `IfNotPresent`                                          |
| `replicas`               | Deployment replica count.                               | `1`                                                     |
| `watchNamespaces`        | Namespaces to watch. Empty means cluster-wide.          | `[]`                                                    |
| `leaderElection.enabled` | Enable controller leader election.                      | `true`                                                  |

## Pod Values

| Key                  | Description                                                 | Default                              |
| -------------------- | ----------------------------------------------------------- | ------------------------------------ |
| `resources`          | Container resource requests and limits.                     | See `values.yaml`                    |
| `sidecars`           | Additional containers added to the Pod.                     | `[]`                                 |
| `podAnnotations`     | Extra Pod annotations.                                      | `{}`                                 |
| `podLabels`          | Extra Pod labels.                                           | `{}`                                 |
| `nodeSelector`       | Pod node selector.                                          | `{}`                                 |
| `tolerations`        | Pod tolerations.                                            | `[]`                                 |
| `affinity`           | Pod affinity rules.                                         | `{}`                                 |
| `podSecurityContext` | Pod-level security context.                                 | `{}`                                 |
| `securityContext`    | Container security context.                                 | `{}`                                 |
| `env`                | Extra environment variables for the exporter container.     | `[{name: TZ, value: Europe/Zurich}]` |
| `extraArgs`          | Additional CLI arguments appended to the container command. | `[]`                                 |

## Probe Values

| Key                      | Description                 | Default           |
| ------------------------ | --------------------------- | ----------------- |
| `startupProbe.enabled`   | Enable the startup probe.   | `true`            |
| `startupProbe.spec`      | Startup probe definition.   | See `values.yaml` |
| `livenessProbe.enabled`  | Enable the liveness probe.  | `true`            |
| `livenessProbe.spec`     | Liveness probe definition.  | See `values.yaml` |
| `readinessProbe.enabled` | Enable the readiness probe. | `true`            |
| `readinessProbe.spec`    | Readiness probe definition. | See `values.yaml` |

## Metrics Values

| Key                                       | Description                                                                   | Default           |
| ----------------------------------------- | ----------------------------------------------------------------------------- | ----------------- |
| `metrics.enabled`                         | Enable the metrics endpoint.                                                  | `true`            |
| `metrics.address`                         | Override the metrics bind address when set.                                   | unset             |
| `metrics.secure`                          | Serve metrics over HTTPS and configure the ServiceMonitor scheme accordingly. | `true`            |
| `metrics.service.enabled`                 | Create a Service for the metrics endpoint.                                    | `true`            |
| `metrics.service.type`                    | Metrics Service type.                                                         | `ClusterIP`       |
| `metrics.service.ports`                   | Metrics Service ports.                                                        | See `values.yaml` |
| `metrics.serviceMonitor.enabled`          | Create a ServiceMonitor for Prometheus Operator.                              | `true`            |
| `metrics.prometheusRule.enabled`          | Create a PrometheusRule.                                                      | `true`            |
| `metrics.prometheusRule.namespace`        | Namespace for the PrometheusRule.                                             | `monitoring`      |
| `metrics.prometheusRule.severity`         | Alert severity label.                                                         | `critical`        |
| `metrics.prometheusRule.additionalLabels` | Additional labels for the PrometheusRule.                                     | `{}`              |

## RBAC Values

| Key                          | Description                                                                     | Default |
| ---------------------------- | ------------------------------------------------------------------------------- | ------- |
| `clusterRole.create`         | Create the ClusterRole and ClusterRoleBinding required for cluster-wide access. | `true`  |
| `clusterRole.name`           | Existing ClusterRole name to bind to when not creating one.                     | `""`    |
| `clusterRole.extraRules`     | Additional RBAC rules to append to the ClusterRole.                             | `[]`    |
| `serviceAccount.create`      | Create a ServiceAccount for the Deployment.                                     | `true`  |
| `serviceAccount.annotations` | ServiceAccount annotations.                                                     | `{}`    |
| `serviceAccount.name`        | Existing ServiceAccount name to use.                                            | `""`    |

## Misc Values

| Key                | Description                                       | Default |
| ------------------ | ------------------------------------------------- | ------- |
| `imagePullSecrets` | Image pull secrets for private registries.        | `[]`    |
| `nameOverride`     | Short name override.                              | `""`    |
| `fullnameOverride` | Full release name override.                       | `""`    |
| `extraObjects`     | Extra Kubernetes objects rendered with the chart. | `[]`    |

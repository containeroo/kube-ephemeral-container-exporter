# kube-ephemeral-container-exporter

[![Go Report Card](https://goreportcard.com/badge/github.com/containeroo/kube-ephemeral-container-exporter?style=flat-square)](https://goreportcard.com/report/github.com/containeroo/kube-ephemeral-container-exporter)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/containeroo/kube-ephemeral-container-exporter)
[![Release](https://img.shields.io/github/release/containeroo/kube-ephemeral-container-exporter.svg?style=flat-square)](https://github.com/containeroo/kube-ephemeral-container-exporter/releases/latest)
[![GitHub tag](https://img.shields.io/github/tag/containeroo/kube-ephemeral-container-exporter.svg?style=flat-square)](https://github.com/containeroo/kube-ephemeral-container-exporter/releases/latest)
[![license](https://img.shields.io/github/license/containeroo/kube-ephemeral-container-exporter.svg?style=flat-square)](LICENSE)

`kube-ephemeral-container-exporter` watches Pods and exposes Prometheus metrics for ephemeral containers. It provides visibility into debug container usage by exporting metrics for ephemeral containers attached to a Pod, the number of ephemeral containers currently running, and per-container state such as running, waiting, terminated, and restart count.

## Prerequisites

- A Kubernetes cluster with ephemeral containers enabled and available.
- RBAC permissions to watch and read Pods in the configured namespaces.
- Prometheus or another scraper configured to scrape the exporter metrics endpoint.

## Installation and Usage

- **Helm**: install `containeroo/kube-ephemeral-container-exporter` from `https://charts.containeroo.ch`. The chart source is maintained in [containeroo/helm-charts](https://github.com/containeroo/helm-charts/tree/master/charts/kube-ephemeral-container-exporter).
- **Kustomize/manifests**: apply `deploy/kubernetes/kustomization.yaml` (or the rendered manifests) after setting image/tag/args.

### Namespaced Mode

By default, `kube-ephemeral-container-exporter` watches all namespaces. To restrict it to specific namespaces, pass the `--watch-namespace` flag. This flag can be repeated or comma-separated to specify multiple namespaces. When set, `kube-ephemeral-container-exporter` will only monitor Pods within those namespaces.

If running in namespaced mode, ensure the associated `Role` and `RoleBinding` are configured accordingly. You can use your existing namespaced RBAC manifests as a starting point for custom definitions.

## How It Works

- The exporter watches Pod events.
- It reconciles only when ephemeral-container related Pod fields change.
- It reads attached ephemeral containers from `spec.ephemeralContainers`.
- It reads runtime state from `status.ephemeralContainerStatuses`.
- For matching Pods, it exports metrics describing:
  - whether the Pod has ephemeral containers attached
  - how many ephemeral containers are attached
  - whether the Pod currently has running ephemeral containers
  - how many ephemeral containers are currently running
  - which ephemeral containers exist
  - whether each ephemeral container is running, waiting, or terminated
  - the restart count of each ephemeral container

When labels relevant to a metric series change, such as `node`, `owner_kind`, or `owner_name`, the exporter removes the old series and recreates them with the updated label set.

### Attached vs Running

Ephemeral containers are effectively append-only in the Pod spec. That means an ephemeral container can still appear in `spec.ephemeralContainers` even when it is no longer running.

Because of that, the exporter exposes separate metrics for:

- ephemeral containers attached to the Pod
- ephemeral containers currently running in the Pod

This makes it possible to distinguish between historical debug container attachment and currently active debug sessions.

## What Is Monitored

`kube-ephemeral-container-exporter` monitors **Pods** directly.

Ephemeral containers are a Pod-level feature, so the exporter does not watch Deployments, StatefulSets, or DaemonSets directly. If a Pod belongs to a higher-level controller, the exporter resolves the Pod owner and includes owner metadata in the exported metrics.

## Pod example

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo
  namespace: default
spec:
  containers:
    - name: app
      image: nginx:latest
```

An ephemeral container can later be attached to the Pod for debugging, for example via `kubectl debug`. Once attached, the exporter will detect it and expose metrics for that Pod.

## Start Parameters

| Flag/Parameter                | Description                                                             | Default | Env Var                                 |
| :---------------------------- | :---------------------------------------------------------------------- | :------ | :-------------------------------------- |
| `--watch-namespace`           | Namespaces to watch (repeatable/comma-separated). Watches all if unset. | (all)   | `EC_EXPORTER_WATCH_NAMESPACE`           |
| `--metrics-enabled`           | Enable or disable the metrics endpoint.                                 | `true`  | `EC_EXPORTER_METRICS_ENABLED`           |
| `--metrics-bind-address`      | Metrics server address (e.g. `:8443`).                                  | `:8443` | `EC_EXPORTER_METRICS_BIND_ADDRESS`      |
| `--metrics-secure`            | Serve metrics over HTTPS.                                               | `true`  | `EC_EXPORTER_METRICS_SECURE`            |
| `--enable-http2`              | Enable HTTP/2 for servers.                                              | `false` | `EC_EXPORTER_ENABLE_HTTP2`              |
| `--health-probe-bind-address` | Health and readiness probe address.                                     | `:8081` | `EC_EXPORTER_HEALTH_PROBE_BIND_ADDRESS` |
| `--leader-elect`              | Enable leader election.                                                 | `true`  | `EC_EXPORTER_LEADER_ELECT`              |
| `--log-encoder`               | Log format (`json`, `console`).                                         | `json`  | `EC_EXPORTER_LOG_ENCODER`               |
| `--log-stacktrace-level`      | Stacktrace log level (`info`, `error`, `panic`).                        | `panic` | `EC_EXPORTER_LOG_STACKTRACE_LEVEL`      |
| `--log-devel`                 | Enable development mode logging.                                        | `false` | `EC_EXPORTER_LOG_DEVEL`                 |

## Prometheus Metrics

`kube-ephemeral-container-exporter` exposes pod-level and container-level metrics for ephemeral containers.

### Available Metrics

1. **Pod Has Ephemeral Containers Attached**
   - **Metric:** `kube_pod_ephemeral_container_present`
   - **Labels:** `namespace`, `pod`, `node`, `owner_kind`, `owner_name`

2. **Attached Ephemeral Container Count**
   - **Metric:** `kube_pod_ephemeral_container_count`
   - **Labels:** `namespace`, `pod`, `node`, `owner_kind`, `owner_name`

3. **Pod Has Running Ephemeral Containers**
   - **Metric:** `kube_pod_ephemeral_container_running_present`
   - **Labels:** `namespace`, `pod`, `node`, `owner_kind`, `owner_name`

4. **Running Ephemeral Container Count**
   - **Metric:** `kube_pod_ephemeral_container_running_count`
   - **Labels:** `namespace`, `pod`, `node`, `owner_kind`, `owner_name`

5. **Ephemeral Container Info**
   - **Metric:** `kube_pod_ephemeral_container_info`
   - **Labels:** `namespace`, `pod`, `node`, `owner_kind`, `owner_name`, `container`, `image`

6. **Ephemeral Container Running**
   - **Metric:** `kube_pod_ephemeral_container_running`
   - **Labels:** `namespace`, `pod`, `container`

7. **Ephemeral Container Terminated**
   - **Metric:** `kube_pod_ephemeral_container_terminated`
   - **Labels:** `namespace`, `pod`, `container`

8. **Ephemeral Container Waiting**
   - **Metric:** `kube_pod_ephemeral_container_waiting`
   - **Labels:** `namespace`, `pod`, `container`

9. **Ephemeral Container Restart Count**
   - **Metric:** `kube_pod_ephemeral_container_restart_count`
   - **Labels:** `namespace`, `pod`, `container`

## Notes on Labels

- `owner_kind` and `owner_name` refer to the direct controller owner of the Pod when one exists.
- `node` is included as a metric label, so when a Pod changes nodes, old metric series are removed and recreated with the new node label.
- `image` can increase cardinality depending on cluster usage patterns.
- Pod-level metrics with `*_present` indicate whether at least one matching ephemeral container exists for the Pod.
- Pod-level metrics with `*_count` expose the corresponding count for the Pod.

## Running locally

```bash
GOCACHE=$(pwd)/.cache/go-build go run ./cmd/main.go
```

To restrict the exporter to one namespace while running locally:

```bash
GOCACHE=$(pwd)/.cache/go-build go run ./cmd/main.go \
  --watch-namespace default
```

## Testing

- Unit tests: `GOCACHE=$(pwd)/.cache/go-build go test ./...`

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.

# kube-ephemeral-container-exporter

[![Go Report Card](https://goreportcard.com/badge/github.com/containeroo/kube-ephemeral-container-exporter?style=flat-square)](https://goreportcard.com/report/github.com/containeroo/kube-ephemeral-container-exporter)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/containeroo/kube-ephemeral-container-exporter)
[![Release](https://img.shields.io/github/release/containeroo/kube-ephemeral-container-exporter.svg?style=flat-square)](https://github.com/containeroo/kube-ephemeral-container-exporter/releases/latest)
[![GitHub tag](https://img.shields.io/github/tag/containeroo/kube-ephemeral-container-exporter.svg?style=flat-square)](https://github.com/containeroo/kube-ephemeral-container-exporter/releases/latest)
[![license](https://img.shields.io/github/license/containeroo/kube-ephemeral-container-exporter.svg?style=flat-square)](LICENSE)

`kube-ephemeral-container-exporter` watches Pods and exposes Prometheus metrics whenever ephemeral containers are attached to them. It is intended to provide visibility into debug container usage by exporting pod-level and container-level metrics for ephemeral containers and their states.

## Prerequisites

- A Kubernetes cluster with ephemeral containers enabled and available.
- RBAC permissions to watch and read Pods in the configured namespaces.
- Prometheus or another scraper configured to scrape the exporter metrics endpoint.

## Installation and Usage

- **Helm**: `helm upgrade --install kube-ephemeral-container-exporter ./deploy/kubernetes/chart/kube-ephemeral-container-exporter`
- **Kustomize/manifests**: apply `deploy/kubernetes/kustomization.yaml` (or the rendered manifests) after setting image/tag/args.

### Namespaced Mode

By default, `kube-ephemeral-container-exporter` watches all namespaces. To restrict it to specific namespaces, pass the `--watch-namespace` flag. This flag can be repeated or comma-separated to specify multiple namespaces. When set, `kube-ephemeral-container-exporter` will only monitor Pods within those namespaces.

If running in namespaced mode, ensure the associated `Role` and `RoleBinding` are configured accordingly. You can use your existing namespaced RBAC manifests as a starting point for custom definitions.

## How It Works

- The exporter watches Pod events.
- It reconciles only when ephemeral-container related Pod fields change.
- For matching Pods, it exports metrics describing:
  - whether the Pod currently has ephemeral containers attached
  - how many ephemeral containers are attached
  - which ephemeral containers exist
  - whether they are running, waiting, or terminated
  - their restart count

- When labels relevant to the metric series change, the exporter removes old metric series and recreates them with the updated label set.

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

1. **Pod Has Ephemeral Containers**
   - **Metric:** `kube_pod_ephemeral_container_present`
   - **Labels:** `namespace`, `pod`, `node`, `owner_kind`, `owner_name`

2. **Ephemeral Container Count**
   - **Metric:** `kube_pod_ephemeral_container_count`
   - **Labels:** `namespace`, `pod`, `node`, `owner_kind`, `owner_name`

3. **Ephemeral Container Info**
   - **Metric:** `kube_pod_ephemeral_container_info`
   - **Labels:** `namespace`, `pod`, `node`, `owner_kind`, `owner_name`, `container`, `image`

4. **Ephemeral Container Running**
   - **Metric:** `kube_pod_ephemeral_container_running`
   - **Labels:** `namespace`, `pod`, `container`

5. **Ephemeral Container Terminated**
   - **Metric:** `kube_pod_ephemeral_container_terminated`
   - **Labels:** `namespace`, `pod`, `container`, `reason`

6. **Ephemeral Container Waiting**
   - **Metric:** `kube_pod_ephemeral_container_waiting`
   - **Labels:** `namespace`, `pod`, `container`, `reason`

7. **Ephemeral Container Restart Count**
   - **Metric:** `kube_pod_ephemeral_container_restart_count`
   - **Labels:** `namespace`, `pod`, `container`

## Notes on Labels

- `owner_kind` and `owner_name` refer to the direct controller owner of the Pod when one exists.
- `node` is included as a metric label, so when a Pod changes nodes, old metric series are removed and recreated with the new node label.
- Labels such as `image` and `reason` can increase cardinality depending on cluster usage patterns.

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

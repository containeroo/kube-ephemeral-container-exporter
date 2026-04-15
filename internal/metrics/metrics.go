/*
Copyright 2026 containeroo.ch

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/prometheus/client_golang/prometheus"
)

// Registry holds Prometheus metrics for the exporter.
type Registry struct {
	podPresent            *prometheus.GaugeVec
	podCount              *prometheus.GaugeVec
	podRunningPresent     *prometheus.GaugeVec
	podRunningCount       *prometheus.GaugeVec
	containerInfo         *prometheus.GaugeVec
	containerRunning      *prometheus.GaugeVec
	containerTerminated   *prometheus.GaugeVec
	containerWaiting      *prometheus.GaugeVec
	containerRestartCount *prometheus.GaugeVec
}

// NewRegistry builds a new Prometheus metrics registry facade and registers all
// collectors on the provided registerer. If reg is nil, the default registerer
// is used.
func NewRegistry(reg prometheus.Registerer) *Registry {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	podPresent := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_pod_ephemeral_container_present",
			Help: "Whether a pod has at least one ephemeral container attached in spec (1 = yes, 0 = no)",
		},
		[]string{"namespace", "pod", "node", "owner_kind", "owner_name"},
	)

	podCount := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_pod_ephemeral_container_count",
			Help: "Number of ephemeral containers attached to a pod in spec",
		},
		[]string{"namespace", "pod", "node", "owner_kind", "owner_name"},
	)

	podRunningPresent := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_pod_ephemeral_container_running_present",
			Help: "Whether a pod currently has at least one running ephemeral container (1 = yes, 0 = no)",
		},
		[]string{"namespace", "pod", "node", "owner_kind", "owner_name"},
	)

	podRunningCount := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_pod_ephemeral_container_running_count",
			Help: "Number of currently running ephemeral containers in a pod",
		},
		[]string{"namespace", "pod", "node", "owner_kind", "owner_name"},
	)

	containerInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_pod_ephemeral_container_info",
			Help: "Metadata about an ephemeral container attached to a pod",
		},
		[]string{"namespace", "pod", "node", "owner_kind", "owner_name", "container", "image"},
	)

	containerRunning := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_pod_ephemeral_container_running",
			Help: "Whether an ephemeral container is currently running (1 = yes, 0 = no)",
		},
		[]string{"namespace", "pod", "container"},
	)

	containerTerminated := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_pod_ephemeral_container_terminated",
			Help: "Whether an ephemeral container is currently terminated (1 = yes, 0 = no)",
		},
		[]string{"namespace", "pod", "container"},
	)

	containerWaiting := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_pod_ephemeral_container_waiting",
			Help: "Whether an ephemeral container is currently waiting (1 = yes, 0 = no)",
		},
		[]string{"namespace", "pod", "container"},
	)

	containerRestartCount := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_pod_ephemeral_container_restart_count",
			Help: "Restart count of an ephemeral container",
		},
		[]string{"namespace", "pod", "container"},
	)

	reg.MustRegister(
		podPresent,
		podCount,
		podRunningPresent,
		podRunningCount,
		containerInfo,
		containerRunning,
		containerTerminated,
		containerWaiting,
		containerRestartCount,
	)

	return &Registry{
		podPresent:            podPresent,
		podCount:              podCount,
		podRunningPresent:     podRunningPresent,
		podRunningCount:       podRunningCount,
		containerInfo:         containerInfo,
		containerRunning:      containerRunning,
		containerTerminated:   containerTerminated,
		containerWaiting:      containerWaiting,
		containerRestartCount: containerRestartCount,
	}
}

// PodPresent exposes the pod attached-presence metric.
func (r *Registry) PodPresent() *prometheus.GaugeVec {
	return r.podPresent
}

// PodCount exposes the pod attached-count metric.
func (r *Registry) PodCount() *prometheus.GaugeVec {
	return r.podCount
}

// PodRunningPresent exposes the pod running-presence metric.
func (r *Registry) PodRunningPresent() *prometheus.GaugeVec {
	return r.podRunningPresent
}

// PodRunningCount exposes the pod running-count metric.
func (r *Registry) PodRunningCount() *prometheus.GaugeVec {
	return r.podRunningCount
}

// ContainerRunning exposes the container running metric.
func (r *Registry) ContainerRunning() *prometheus.GaugeVec {
	return r.containerRunning
}

// UpdatePodEphemeralContainers updates all ephemeral-container metrics for a pod.
func (r *Registry) UpdatePodEphemeralContainers(pod *corev1.Pod, ownerKind, ownerName string) {
	namespace := pod.Namespace
	podName := pod.Name
	node := pod.Spec.NodeName

	attachedCount := len(pod.Spec.EphemeralContainers)
	attachedPresent := boolToFloat(attachedCount > 0)

	r.podPresent.WithLabelValues(namespace, podName, node, ownerKind, ownerName).Set(attachedPresent)
	r.podCount.WithLabelValues(namespace, podName, node, ownerKind, ownerName).Set(float64(attachedCount))

	for _, container := range pod.Spec.EphemeralContainers {
		r.containerInfo.WithLabelValues(
			namespace,
			podName,
			node,
			ownerKind,
			ownerName,
			container.Name,
			container.Image,
		).Set(1)
	}

	runningCount := 0

	for _, status := range pod.Status.EphemeralContainerStatuses {
		isRunning := status.State.Running != nil
		isTerminated := status.State.Terminated != nil
		isWaiting := status.State.Waiting != nil

		if isRunning {
			runningCount++
		}

		r.containerRunning.WithLabelValues(namespace, podName, status.Name).Set(boolToFloat(isRunning))
		r.containerTerminated.WithLabelValues(namespace, podName, status.Name).Set(boolToFloat(isTerminated))
		r.containerWaiting.WithLabelValues(namespace, podName, status.Name).Set(boolToFloat(isWaiting))
		r.containerRestartCount.WithLabelValues(namespace, podName, status.Name).Set(float64(status.RestartCount))
	}

	r.podRunningPresent.WithLabelValues(namespace, podName, node, ownerKind, ownerName).
		Set(boolToFloat(runningCount > 0))
	r.podRunningCount.WithLabelValues(namespace, podName, node, ownerKind, ownerName).
		Set(float64(runningCount))
}

// DeletePodEphemeralContainers removes all metrics for the pod that can be derived from the given object.
func (r *Registry) DeletePodEphemeralContainers(pod *corev1.Pod, ownerKind, ownerName string) {
	namespace := pod.Namespace
	podName := pod.Name
	node := pod.Spec.NodeName

	r.podPresent.DeleteLabelValues(namespace, podName, node, ownerKind, ownerName)
	r.podCount.DeleteLabelValues(namespace, podName, node, ownerKind, ownerName)
	r.podRunningPresent.DeleteLabelValues(namespace, podName, node, ownerKind, ownerName)
	r.podRunningCount.DeleteLabelValues(namespace, podName, node, ownerKind, ownerName)

	for _, container := range pod.Spec.EphemeralContainers {
		r.containerInfo.DeleteLabelValues(
			namespace,
			podName,
			node,
			ownerKind,
			ownerName,
			container.Name,
			container.Image,
		)
	}

	for _, status := range pod.Status.EphemeralContainerStatuses {
		r.containerRunning.DeleteLabelValues(namespace, podName, status.Name)
		r.containerTerminated.DeleteLabelValues(namespace, podName, status.Name)
		r.containerWaiting.DeleteLabelValues(namespace, podName, status.Name)
		r.containerRestartCount.DeleteLabelValues(namespace, podName, status.Name)
	}
}

// boolToFloat converts a bool to a Prometheus-friendly numeric value.
func boolToFloat(v bool) float64 {
	if v {
		return 1
	}
	return 0
}

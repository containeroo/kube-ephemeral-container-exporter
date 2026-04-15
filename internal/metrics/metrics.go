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

// Registry holds Prometheus metrics for the operator.
type Registry struct {
	reg                   prometheus.Registerer
	podPresent            *prometheus.GaugeVec
	podCount              *prometheus.GaugeVec
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
			Help: "Whether a pod currently has at least one ephemeral container attached (1 = yes, 0 = no)",
		},
		[]string{"namespace", "pod", "node", "owner_kind", "owner_name"},
	)

	podCount := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_pod_ephemeral_container_count",
			Help: "Number of ephemeral containers currently attached to a pod",
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
		[]string{"namespace", "pod", "container", "reason"},
	)

	containerWaiting := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kube_pod_ephemeral_container_waiting",
			Help: "Whether an ephemeral container is currently waiting (1 = yes, 0 = no)",
		},
		[]string{"namespace", "pod", "container", "reason"},
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
		containerInfo,
		containerRunning,
		containerTerminated,
		containerWaiting,
		containerRestartCount,
	)

	return &Registry{
		reg:                   reg,
		podPresent:            podPresent,
		podCount:              podCount,
		containerInfo:         containerInfo,
		containerRunning:      containerRunning,
		containerTerminated:   containerTerminated,
		containerWaiting:      containerWaiting,
		containerRestartCount: containerRestartCount,
	}
}

// UpdatePodEphemeralContainers updates all ephemeral-container metrics for a pod.
func (r *Registry) UpdatePodEphemeralContainers(pod *corev1.Pod, ownerKind, ownerName string) {
	namespace := pod.Namespace
	podName := pod.Name
	node := pod.Spec.NodeName
	count := len(pod.Spec.EphemeralContainers)

	present := 0.0
	if count > 0 {
		present = 1.0
	}

	r.podPresent.WithLabelValues(namespace, podName, node, ownerKind, ownerName).Set(present)
	r.podCount.WithLabelValues(namespace, podName, node, ownerKind, ownerName).Set(float64(count))

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

	for _, status := range pod.Status.EphemeralContainerStatuses {
		r.containerRunning.WithLabelValues(namespace, podName, status.Name).
			Set(boolToFloat(status.State.Running != nil))

		r.containerRestartCount.WithLabelValues(namespace, podName, status.Name).
			Set(float64(status.RestartCount))

		terminatedReason := ""
		if status.State.Terminated != nil {
			terminatedReason = status.State.Terminated.Reason
		}
		r.containerTerminated.WithLabelValues(namespace, podName, status.Name, terminatedReason).
			Set(boolToFloat(status.State.Terminated != nil))

		waitingReason := ""
		if status.State.Waiting != nil {
			waitingReason = status.State.Waiting.Reason
		}
		r.containerWaiting.WithLabelValues(namespace, podName, status.Name, waitingReason).
			Set(boolToFloat(status.State.Waiting != nil))
	}
}

// DeletePodEphemeralContainers removes all metrics for the pod that can be derived from the given object.
func (r *Registry) DeletePodEphemeralContainers(pod *corev1.Pod, ownerKind, ownerName string) {
	namespace := pod.Namespace
	podName := pod.Name
	node := pod.Spec.NodeName

	r.podPresent.DeleteLabelValues(namespace, podName, node, ownerKind, ownerName)
	r.podCount.DeleteLabelValues(namespace, podName, node, ownerKind, ownerName)

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
		r.containerRestartCount.DeleteLabelValues(namespace, podName, status.Name)

		terminatedReason := ""
		if status.State.Terminated != nil {
			terminatedReason = status.State.Terminated.Reason
		}
		r.containerTerminated.DeleteLabelValues(namespace, podName, status.Name, terminatedReason)

		waitingReason := ""
		if status.State.Waiting != nil {
			waitingReason = status.State.Waiting.Reason
		}
		r.containerWaiting.DeleteLabelValues(namespace, podName, status.Name, waitingReason)
	}
}

// boolToFloat converts a bool to a Prometheus-friendly numeric value.
func boolToFloat(v bool) float64 {
	if v {
		return 1
	}
	return 0
}

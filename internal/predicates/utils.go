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

package predicates

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
)

// exportedEphemeralContainerSpec is the metric-relevant subset of an
// ephemeral-container spec used for predicate comparisons.
type exportedEphemeralContainerSpec struct {
	name  string
	image string
}

// exportedEphemeralContainerStatus is the metric-relevant subset of an
// ephemeral-container status used for predicate comparisons.
type exportedEphemeralContainerStatus struct {
	name         string
	restartCount int32
	running      bool
	terminated   bool
	waiting      bool
}

// ephemeralContainersChanged reports whether metric-relevant ephemeral container
// spec fields changed.
func ephemeralContainersChanged(oldPod, newPod *corev1.Pod) bool {
	return !reflect.DeepEqual(
		exportedEphemeralContainerSpecs(oldPod.Spec.EphemeralContainers),
		exportedEphemeralContainerSpecs(newPod.Spec.EphemeralContainers),
	)
}

// ephemeralContainerStatusesChanged reports whether metric-relevant ephemeral
// container status fields changed.
func ephemeralContainerStatusesChanged(oldPod, newPod *corev1.Pod) bool {
	return !reflect.DeepEqual(
		exportedEphemeralContainerStatuses(oldPod.Status.EphemeralContainerStatuses),
		exportedEphemeralContainerStatuses(newPod.Status.EphemeralContainerStatuses),
	)
}

// podNodeChanged reports whether the Pod's node name changed, requiring old
// metrics to be deleted and recreated with the new node label.
func podNodeChanged(oldPod, newPod *corev1.Pod) bool {
	return oldPod.Spec.NodeName != newPod.Spec.NodeName
}

// podOwnerChanged reports whether the Pod's owner references changed.
func podOwnerChanged(oldPod, newPod *corev1.Pod) bool {
	return !reflect.DeepEqual(oldPod.OwnerReferences, newPod.OwnerReferences)
}

// exportedEphemeralContainerSpecs reduces ephemeral-container specs to the
// fields that affect exported metrics.
func exportedEphemeralContainerSpecs(containers []corev1.EphemeralContainer) []exportedEphemeralContainerSpec {
	specs := make([]exportedEphemeralContainerSpec, 0, len(containers))
	for _, container := range containers {
		specs = append(specs, exportedEphemeralContainerSpec{
			name:  container.Name,
			image: container.Image,
		})
	}

	return specs
}

// exportedEphemeralContainerStatuses reduces container statuses to the fields
// that affect exported metrics.
func exportedEphemeralContainerStatuses(statuses []corev1.ContainerStatus) []exportedEphemeralContainerStatus {
	exported := make([]exportedEphemeralContainerStatus, 0, len(statuses))
	for _, status := range statuses {
		exported = append(exported, exportedEphemeralContainerStatus{
			name:         status.Name,
			restartCount: status.RestartCount,
			running:      status.State.Running != nil,
			terminated:   status.State.Terminated != nil,
			waiting:      status.State.Waiting != nil,
		})
	}

	return exported
}

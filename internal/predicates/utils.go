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

// ephemeralContainersChanged reports whether the Pod's ephemeral container spec changed.
func ephemeralContainersChanged(oldPod, newPod *corev1.Pod) bool {
	return !reflect.DeepEqual(oldPod.Spec.EphemeralContainers, newPod.Spec.EphemeralContainers)
}

// ephemeralContainerStatusesChanged reports whether the Pod's ephemeral container statuses changed.
func ephemeralContainerStatusesChanged(oldPod, newPod *corev1.Pod) bool {
	return !reflect.DeepEqual(oldPod.Status.EphemeralContainerStatuses, newPod.Status.EphemeralContainerStatuses)
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

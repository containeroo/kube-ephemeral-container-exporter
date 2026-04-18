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
	"github.com/containeroo/kube-ephemeral-container-exporter/internal/metrics"
	"github.com/containeroo/kube-ephemeral-container-exporter/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// EphemeralContainerChanges filters Pod events to reconcile only when
// ephemeral-container related state changes or when cleanup is needed.
func EphemeralContainerChanges(reg *metrics.Registry) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			pod, ok := e.Object.(*corev1.Pod)
			if !ok {
				return false
			}
			return utils.PodHasEphemeralContainers(pod)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldPod, okOld := e.ObjectOld.(*corev1.Pod)
			newPod, okNew := e.ObjectNew.(*corev1.Pod)
			if !okOld || !okNew {
				return false
			}

			// Ignore updates for Pods that never had ephemeral containers and still
			// do not have any.
			if !utils.PodHasEphemeralContainers(oldPod) && !utils.PodHasEphemeralContainers(newPod) {
				return false
			}

			// Metric-relevant status changes should reconcile so
			// running/waiting/terminated/restart metrics stay current.
			if ephemeralContainerStatusesChanged(oldPod, newPod) {
				return true
			}

			// Spec/node/owner changes also reconcile so the metrics registry can
			// rebuild the pod's derived series from the latest labels and spec.
			if ephemeralContainersChanged(oldPod, newPod) ||
				podNodeChanged(oldPod, newPod) ||
				podOwnerChanged(oldPod, newPod) {
				return true
			}

			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			pod, ok := e.Object.(*corev1.Pod)
			if !ok {
				return false
			}

			// Delete events cannot be reconciled from the API anymore, so cleanup
			// the pod's derived metric series directly and do not enqueue.
			reg.DeletePodEphemeralContainers(pod)
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			pod, ok := e.Object.(*corev1.Pod)
			if !ok {
				return false
			}
			return utils.PodHasEphemeralContainers(pod)
		},
	}
}

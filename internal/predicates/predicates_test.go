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
	"testing"

	"github.com/containeroo/kube-ephemeral-container-exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestEphemeralContainerChanges(t *testing.T) {
	t.Parallel()

	promReg := prometheus.NewRegistry()
	reg := metrics.NewRegistry(promReg)
	pred := EphemeralContainerChanges(reg)

	podWithoutEphemeral := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "plain",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			NodeName: "node-a",
		},
	}

	podWithEphemeral := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "debug",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "debug-rs",
					Controller: boolPtr(true),
				},
			},
		},
		Spec: corev1.PodSpec{
			NodeName: "node-a",
			EphemeralContainers: []corev1.EphemeralContainer{
				{
					EphemeralContainerCommon: corev1.EphemeralContainerCommon{
						Name:  "debugger",
						Image: "busybox:latest",
					},
				},
			},
		},
		Status: corev1.PodStatus{
			EphemeralContainerStatuses: []corev1.ContainerStatus{
				{
					Name: "debugger",
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{},
					},
				},
			},
		},
	}

	t.Run("Create event allowed when ephemeral container exists", func(t *testing.T) {
		e := event.CreateEvent{Object: podWithEphemeral}
		assert.True(t, pred.Create(e))
	})

	t.Run("Create event denied when ephemeral container missing", func(t *testing.T) {
		e := event.CreateEvent{Object: podWithoutEphemeral}
		assert.False(t, pred.Create(e))
	})

	t.Run("Update event allowed when ephemeral containers changed", func(t *testing.T) {
		oldPod := podWithoutEphemeral.DeepCopy()
		newPod := podWithEphemeral.DeepCopy()

		e := event.UpdateEvent{ObjectOld: oldPod, ObjectNew: newPod}
		assert.True(t, pred.Update(e))
	})

	t.Run("Update event allowed when ephemeral container statuses changed", func(t *testing.T) {
		oldPod := podWithEphemeral.DeepCopy()
		newPod := podWithEphemeral.DeepCopy()
		newPod.Status.EphemeralContainerStatuses = []corev1.ContainerStatus{
			{
				Name: "debugger",
				State: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{
						Reason: "Completed",
					},
				},
			},
		}

		e := event.UpdateEvent{ObjectOld: oldPod, ObjectNew: newPod}
		assert.True(t, pred.Update(e))
	})

	t.Run("Update event denied when only non-exported status fields changed", func(t *testing.T) {
		oldPod := podWithEphemeral.DeepCopy()
		newPod := podWithEphemeral.DeepCopy()
		newPod.Status.EphemeralContainerStatuses[0].ContainerID = "containerd://debugger"
		newPod.Status.EphemeralContainerStatuses[0].ImageID = "sha256:deadbeef"

		e := event.UpdateEvent{ObjectOld: oldPod, ObjectNew: newPod}
		assert.False(t, pred.Update(e))
	})

	t.Run("Update event allowed when node changed", func(t *testing.T) {
		oldPod := podWithEphemeral.DeepCopy()
		newPod := podWithEphemeral.DeepCopy()
		newPod.Spec.NodeName = "node-b"

		e := event.UpdateEvent{ObjectOld: oldPod, ObjectNew: newPod}
		assert.True(t, pred.Update(e))
	})

	t.Run("Update event allowed when owner changed", func(t *testing.T) {
		oldPod := podWithEphemeral.DeepCopy()
		newPod := podWithEphemeral.DeepCopy()
		newPod.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: "apps/v1",
				Kind:       "ReplicaSet",
				Name:       "other-rs",
				Controller: boolPtr(true),
			},
		}

		e := event.UpdateEvent{ObjectOld: oldPod, ObjectNew: newPod}
		assert.True(t, pred.Update(e))
	})

	t.Run("Update event denied when nothing relevant changed", func(t *testing.T) {
		oldPod := podWithEphemeral.DeepCopy()
		newPod := podWithEphemeral.DeepCopy()

		e := event.UpdateEvent{ObjectOld: oldPod, ObjectNew: newPod}
		assert.False(t, pred.Update(e))
	})

	t.Run("Update event denied when neither old nor new pod has ephemeral containers", func(t *testing.T) {
		oldPod := podWithoutEphemeral.DeepCopy()
		newPod := podWithoutEphemeral.DeepCopy()
		newPod.Spec.NodeName = "node-b"

		e := event.UpdateEvent{ObjectOld: oldPod, ObjectNew: newPod}
		assert.False(t, pred.Update(e))
	})

	t.Run("Delete event handled and not queued", func(t *testing.T) {
		e := event.DeleteEvent{Object: podWithEphemeral}
		assert.False(t, pred.Delete(e))
	})

	t.Run("Generic event allowed when ephemeral container exists", func(t *testing.T) {
		e := event.GenericEvent{Object: podWithEphemeral}
		assert.True(t, pred.Generic(e))
	})

	t.Run("Generic event denied when ephemeral container missing", func(t *testing.T) {
		e := event.GenericEvent{Object: podWithoutEphemeral}
		assert.False(t, pred.Generic(e))
	})
}

func boolPtr(v bool) *bool {
	return &v
}

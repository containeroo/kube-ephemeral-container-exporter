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

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilterFunctions(t *testing.T) {
	t.Parallel()

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

	t.Run("helper ephemeralContainersChanged", func(t *testing.T) {
		t.Parallel()
		assert.True(t, ephemeralContainersChanged(podWithoutEphemeral, podWithEphemeral))
		assert.False(t, ephemeralContainersChanged(podWithEphemeral, podWithEphemeral.DeepCopy()))
	})

	t.Run("helper ephemeralContainerStatusesChanged", func(t *testing.T) {
		t.Parallel()
		oldPod := podWithEphemeral.DeepCopy()
		newPod := podWithEphemeral.DeepCopy()
		newPod.Status.EphemeralContainerStatuses = nil

		assert.True(t, ephemeralContainerStatusesChanged(oldPod, newPod))
		assert.False(t, ephemeralContainerStatusesChanged(oldPod, oldPod.DeepCopy()))
	})

	t.Run("helper podNodeChanged", func(t *testing.T) {
		t.Parallel()
		oldPod := podWithEphemeral.DeepCopy()
		newPod := podWithEphemeral.DeepCopy()
		newPod.Spec.NodeName = "node-b"

		assert.True(t, podNodeChanged(oldPod, newPod))
		assert.False(t, podNodeChanged(oldPod, oldPod.DeepCopy()))
	})

	t.Run("helper podOwnerChanged", func(t *testing.T) {
		t.Parallel()
		oldPod := podWithEphemeral.DeepCopy()
		newPod := podWithEphemeral.DeepCopy()
		newPod.OwnerReferences = nil

		assert.True(t, podOwnerChanged(oldPod, newPod))
		assert.False(t, podOwnerChanged(oldPod, oldPod.DeepCopy()))
	})
}

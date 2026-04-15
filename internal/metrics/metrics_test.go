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
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBoolToFloat(t *testing.T) {
	t.Parallel()

	t.Run("true", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, float64(1), boolToFloat(true))
	})

	t.Run("false", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, float64(0), boolToFloat(false))
	})
}

func TestRegistryMetrics(t *testing.T) {
	t.Parallel()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "demo",
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
				{
					EphemeralContainerCommon: corev1.EphemeralContainerCommon{
						Name:  "shell",
						Image: "alpine:3",
					},
				},
			},
		},
		Status: corev1.PodStatus{
			EphemeralContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "debugger",
					RestartCount: 2,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{},
					},
				},
				{
					Name:         "shell",
					RestartCount: 1,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason: "CrashLoopBackOff",
						},
					},
				},
				{
					Name: "terminated",
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							Reason: "Completed",
						},
					},
				},
			},
		},
	}

	t.Run("UpdatePodEphemeralContainers", func(t *testing.T) {
		t.Parallel()

		promReg := prometheus.NewRegistry()
		reg := NewRegistry(promReg)

		reg.UpdatePodEphemeralContainers(pod, "ReplicaSet", "demo-rs")

		require.Equal(t, float64(1), testutil.ToFloat64(
			reg.podPresent.WithLabelValues("default", "demo", "node-a", "ReplicaSet", "demo-rs"),
		))
		require.Equal(t, float64(2), testutil.ToFloat64(
			reg.podCount.WithLabelValues("default", "demo", "node-a", "ReplicaSet", "demo-rs"),
		))

		require.Equal(t, float64(1), testutil.ToFloat64(
			reg.containerInfo.WithLabelValues("default", "demo", "node-a", "ReplicaSet", "demo-rs", "debugger", "busybox:latest"),
		))
		require.Equal(t, float64(1), testutil.ToFloat64(
			reg.containerInfo.WithLabelValues("default", "demo", "node-a", "ReplicaSet", "demo-rs", "shell", "alpine:3"),
		))

		require.Equal(t, float64(1), testutil.ToFloat64(
			reg.containerRunning.WithLabelValues("default", "demo", "debugger"),
		))
		require.Equal(t, float64(2), testutil.ToFloat64(
			reg.containerRestartCount.WithLabelValues("default", "demo", "debugger"),
		))

		require.Equal(t, float64(1), testutil.ToFloat64(
			reg.containerWaiting.WithLabelValues("default", "demo", "shell", "CrashLoopBackOff"),
		))
		require.Equal(t, float64(1), testutil.ToFloat64(
			reg.containerRestartCount.WithLabelValues("default", "demo", "shell"),
		))

		require.Equal(t, float64(1), testutil.ToFloat64(
			reg.containerTerminated.WithLabelValues("default", "demo", "terminated", "Completed"),
		))
	})

	t.Run("DeletePodEphemeralContainers", func(t *testing.T) {
		t.Parallel()

		promReg := prometheus.NewRegistry()
		reg := NewRegistry(promReg)

		reg.UpdatePodEphemeralContainers(pod, "ReplicaSet", "demo-rs")
		reg.DeletePodEphemeralContainers(pod, "ReplicaSet", "demo-rs")

		require.Equal(t, float64(0), testutil.ToFloat64(
			reg.podPresent.WithLabelValues("default", "demo", "node-a", "ReplicaSet", "demo-rs"),
		))
		require.Equal(t, float64(0), testutil.ToFloat64(
			reg.podCount.WithLabelValues("default", "demo", "node-a", "ReplicaSet", "demo-rs"),
		))

		require.Equal(t, float64(0), testutil.ToFloat64(
			reg.containerInfo.WithLabelValues("default", "demo", "node-a", "ReplicaSet", "demo-rs", "debugger", "busybox:latest"),
		))
		require.Equal(t, float64(0), testutil.ToFloat64(
			reg.containerRunning.WithLabelValues("default", "demo", "debugger"),
		))
		require.Equal(t, float64(0), testutil.ToFloat64(
			reg.containerRestartCount.WithLabelValues("default", "demo", "debugger"),
		))
		require.Equal(t, float64(0), testutil.ToFloat64(
			reg.containerWaiting.WithLabelValues("default", "demo", "shell", "CrashLoopBackOff"),
		))
		require.Equal(t, float64(0), testutil.ToFloat64(
			reg.containerTerminated.WithLabelValues("default", "demo", "terminated", "Completed"),
		))
	})

	t.Run("pod without ephemeral containers", func(t *testing.T) {
		t.Parallel()

		promReg := prometheus.NewRegistry()
		reg := NewRegistry(promReg)

		plainPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "plain",
			},
			Spec: corev1.PodSpec{
				NodeName: "node-b",
			},
		}

		reg.UpdatePodEphemeralContainers(plainPod, "", "")

		require.Equal(t, float64(0), testutil.ToFloat64(
			reg.podPresent.WithLabelValues("default", "plain", "node-b", "", ""),
		))
		require.Equal(t, float64(0), testutil.ToFloat64(
			reg.podCount.WithLabelValues("default", "plain", "node-b", "", ""),
		))
	})
}

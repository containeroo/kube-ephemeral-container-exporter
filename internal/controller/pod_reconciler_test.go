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

package controller

import (
	"context"
	"testing"

	"github.com/containeroo/kube-ephemeral-container-exporter/internal/metrics"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPodReconcilerReconcile(t *testing.T) {
	t.Parallel()

	t.Run("pod not found", func(t *testing.T) {
		t.Parallel()

		scheme := newScheme(t)
		kubeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

		promReg := prometheus.NewRegistry()
		metricsReg := metrics.NewRegistry(promReg)

		reconciler := &PodReconciler{
			KubeClient: kubeClient,
			Logger:     logr.Discard(),
			Recorder:   record.NewFakeRecorder(10),
			Metrics:    metricsReg,
		}

		req := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "test-namespace",
				Name:      "missing-pod",
			},
		}

		result, err := reconciler.Reconcile(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)
	})

	t.Run("successful reconcile updates attached and running metrics", func(t *testing.T) {
		t.Parallel()

		controller := true
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "apps/v1",
						Kind:       "ReplicaSet",
						Name:       "test-rs",
						Controller: &controller,
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

		scheme := newScheme(t)
		kubeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(pod).
			Build()

		promReg := prometheus.NewRegistry()
		metricsReg := metrics.NewRegistry(promReg)

		reconciler := &PodReconciler{
			KubeClient: kubeClient,
			Logger:     logr.Discard(),
			Recorder:   record.NewFakeRecorder(10),
			Metrics:    metricsReg,
		}

		req := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "test-namespace",
				Name:      "test-pod",
			},
		}

		result, err := reconciler.Reconcile(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		require.Equal(t, float64(1), testutil.ToFloat64(
			metricsReg.PodPresent().WithLabelValues("test-namespace", "test-pod", "node-a", "ReplicaSet", "test-rs"),
		))
		require.Equal(t, float64(1), testutil.ToFloat64(
			metricsReg.PodCount().WithLabelValues("test-namespace", "test-pod", "node-a", "ReplicaSet", "test-rs"),
		))
		require.Equal(t, float64(1), testutil.ToFloat64(
			metricsReg.PodRunningPresent().WithLabelValues("test-namespace", "test-pod", "node-a", "ReplicaSet", "test-rs"),
		))
		require.Equal(t, float64(1), testutil.ToFloat64(
			metricsReg.PodRunningCount().WithLabelValues("test-namespace", "test-pod", "node-a", "ReplicaSet", "test-rs"),
		))
		require.Equal(t, float64(1), testutil.ToFloat64(
			metricsReg.ContainerRunning().WithLabelValues("test-namespace", "test-pod", "debugger"),
		))
	})

	t.Run("reconcile plain pod writes zero attached and running metrics", func(t *testing.T) {
		t.Parallel()

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "plain-pod",
				Namespace: "test-namespace",
			},
			Spec: corev1.PodSpec{
				NodeName: "node-a",
			},
		}

		scheme := newScheme(t)
		kubeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(pod).
			Build()

		promReg := prometheus.NewRegistry()
		metricsReg := metrics.NewRegistry(promReg)

		reconciler := &PodReconciler{
			KubeClient: kubeClient,
			Logger:     logr.Discard(),
			Recorder:   record.NewFakeRecorder(10),
			Metrics:    metricsReg,
		}

		req := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "test-namespace",
				Name:      "plain-pod",
			},
		}

		result, err := reconciler.Reconcile(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		got := &corev1.Pod{}
		err = kubeClient.Get(context.Background(), types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "plain-pod",
		}, got)
		require.NoError(t, err)
		assert.Equal(t, "plain-pod", got.Name)

		require.Equal(t, float64(0), testutil.ToFloat64(
			metricsReg.PodPresent().WithLabelValues("test-namespace", "plain-pod", "node-a", "", ""),
		))
		require.Equal(t, float64(0), testutil.ToFloat64(
			metricsReg.PodCount().WithLabelValues("test-namespace", "plain-pod", "node-a", "", ""),
		))
		require.Equal(t, float64(0), testutil.ToFloat64(
			metricsReg.PodRunningPresent().WithLabelValues("test-namespace", "plain-pod", "node-a", "", ""),
		))
		require.Equal(t, float64(0), testutil.ToFloat64(
			metricsReg.PodRunningCount().WithLabelValues("test-namespace", "plain-pod", "node-a", "", ""),
		))
	})
}

func TestNewScheme(t *testing.T) {
	t.Parallel()

	scheme := newScheme(t)

	pod := &corev1.Pod{}
	gvks, _, err := scheme.ObjectKinds(pod)
	require.NoError(t, err)
	require.NotEmpty(t, gvks)

	found := false
	for _, gvk := range gvks {
		if gvk.Group == "" && gvk.Version == "v1" && gvk.Kind == "Pod" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestPodReconcilerNotFoundIsIgnored(t *testing.T) {
	t.Parallel()

	scheme := newScheme(t)
	kubeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	promReg := prometheus.NewRegistry()
	metricsReg := metrics.NewRegistry(promReg)

	reconciler := &PodReconciler{
		KubeClient: kubeClient,
		Logger:     logr.Discard(),
		Recorder:   record.NewFakeRecorder(10),
		Metrics:    metricsReg,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "ns1",
			Name:      "does-not-exist",
		},
	}

	_, err := reconciler.Reconcile(context.Background(), req)
	require.NoError(t, err)

	getErr := kubeClient.Get(context.Background(), req.NamespacedName, &corev1.Pod{})
	assert.True(t, apierrors.IsNotFound(getErr))
}

func TestPodReconcilerSetupWithManager(t *testing.T) {
	t.Skip("SetupWithManager is usually covered indirectly; skip unless you want an envtest-based manager test")
}

func newScheme(t *testing.T) *runtime.Scheme {
	t.Helper()

	s := runtime.NewScheme()
	err := corev1.AddToScheme(s)
	require.NoError(t, err)

	return s
}

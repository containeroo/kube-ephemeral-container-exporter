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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

func TestResolvePodOwner(t *testing.T) {
	t.Parallel()
	t.Run("returns direct controller owner", func(t *testing.T) {
		t.Parallel()
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind:       "ReplicaSet",
						Name:       "demo-rs",
						Controller: boolPtr(false),
					},
					{
						Kind:       "Deployment",
						Name:       "demo",
						Controller: boolPtr(true),
					},
				},
			},
		}

		kind, name := ResolvePodOwner(pod)
		assert.Equal(t, "Deployment", kind)
		assert.Equal(t, "demo", name)
	})

	t.Run("returns empty values when no controller owner exists", func(t *testing.T) {
		t.Parallel()
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind:       "ReplicaSet",
						Name:       "demo-rs",
						Controller: boolPtr(false),
					},
				},
			},
		}

		kind, name := ResolvePodOwner(pod)
		assert.Equal(t, "", kind)
		assert.Equal(t, "", name)
	})

	t.Run("returns empty values when owner references are empty", func(t *testing.T) {
		t.Parallel()
		pod := &corev1.Pod{}

		kind, name := ResolvePodOwner(pod)
		assert.Equal(t, "", kind)
		assert.Equal(t, "", name)
	})

	t.Run("returns empty values when controller pointer is nil", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "Deployment",
						Name: "demo",
					},
				},
			},
		}

		kind, name := ResolvePodOwner(pod)
		assert.Equal(t, "", kind)
		assert.Equal(t, "", name)
	})
}

func TestToCacheOptions(t *testing.T) {
	t.Parallel()
	t.Run("returns empty options when no namespaces are given", func(t *testing.T) {
		opts := ToCacheOptions(nil)
		assert.Equal(t, cache.Options{}, opts)
	})

	t.Run("returns options with default namespaces", func(t *testing.T) {
		t.Parallel()
		opts := ToCacheOptions([]string{"default", "kube-system"})

		expected := cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				"default":     {},
				"kube-system": {},
			},
		}

		assert.Equal(t, expected, opts)
	})

	t.Run("deduplicates duplicate namespaces via map keys", func(t *testing.T) {
		opts := ToCacheOptions([]string{"default", "default"})

		expected := cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				"default": {},
			},
		}

		assert.Equal(t, expected, opts)
	})
}

func TestPodHasEphemeralContainers(t *testing.T) {
	t.Parallel()
	t.Run("helper hasEphemeralContainers", func(t *testing.T) {
		t.Parallel()
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
		podWithoutEphemeral := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "plain",
				Namespace: "default",
			},
			Spec: corev1.PodSpec{
				NodeName: "node-a",
			},
		}
		assert.True(t, PodHasEphemeralContainers(podWithEphemeral))
		assert.False(t, PodHasEphemeralContainers(podWithoutEphemeral))
	})
}

func boolPtr(v bool) *bool {
	return &v
}

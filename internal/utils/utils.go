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
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

// ResolvePodOwner returns the direct controller owner of the pod if one exists.
func ResolvePodOwner(pod *corev1.Pod) (string, string) {
	for _, ref := range pod.OwnerReferences {
		if ref.Controller != nil && *ref.Controller {
			return ref.Kind, ref.Name
		}
	}
	return "", ""
}

// ToCacheOptions returns cache.Options configured to watch the given namespaces.
// If no namespaces are provided, it returns an empty Options which watches all namespaces.
func ToCacheOptions(watchNamespaces []string) cache.Options {
	if len(watchNamespaces) == 0 {
		return cache.Options{}
	}

	nsMap := make(map[string]cache.Config, len(watchNamespaces))
	for _, ns := range watchNamespaces {
		nsMap[ns] = cache.Config{}
	}

	return cache.Options{DefaultNamespaces: nsMap}
}

// PodHasEphemeralContainers reports whether the Pod currently has at least one ephemeral container.
func PodHasEphemeralContainers(pod *corev1.Pod) bool {
	return len(pod.Spec.EphemeralContainers) > 0
}

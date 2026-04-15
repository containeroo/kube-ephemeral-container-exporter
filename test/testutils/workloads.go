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

package testutils

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2" // nolint:staticcheck
	. "github.com/onsi/gomega"    // nolint:staticcheck

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultTestImage     string = "nginx:1.29"
	DefaultTestImageName string = "nginx"
)

// CreatePod creates and applies a Pod in the specified namespace.
func CreatePod(ctx context.Context, namespace, name string, opts ...Option) *corev1.Pod {
	meta := metav1.ObjectMeta{
		Name:        name,
		Namespace:   namespace,
		Annotations: map[string]string{},
		Labels:      map[string]string{"app": name},
	}
	spec := corev1.PodSpec{
		Containers: []corev1.Container{
			{Name: DefaultTestImageName, Image: DefaultTestImage},
		},
	}
	pod := &corev1.Pod{ObjectMeta: meta, Spec: spec}
	applyOptions(pod, opts...)
	Expect(K8sClient.Create(ctx, pod)).To(Succeed())

	CheckPodReady(ctx, pod)
	return pod
}

// CreateDeployment creates and applies a Deployment in the specified namespace.
func CreateDeployment(ctx context.Context, namespace, name string, opts ...Option) *appsv1.Deployment {
	meta := metav1.ObjectMeta{
		Name:        name,
		Namespace:   namespace,
		Annotations: map[string]string{},
		Labels:      map[string]string{"app": name},
	}
	spec := appsv1.DeploymentSpec{
		Replicas: Int32Ptr(1),
		Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: DefaultTestImageName, Image: DefaultTestImage},
				},
			},
		},
	}
	deployment := &appsv1.Deployment{ObjectMeta: meta, Spec: spec}
	applyOptions(deployment, opts...)
	Expect(K8sClient.Create(ctx, deployment)).To(Succeed())

	CheckResourceReadiness(ctx, deployment)
	deployment.TypeMeta = metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"}
	return deployment
}

// GetPod fetches a Pod.
func GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	pod := &corev1.Pod{}
	if err := K8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, pod); err != nil {
		return nil, err
	}
	return pod, nil
}

// GetDeploymentPod returns one Pod belonging to the Deployment.
func GetDeploymentPod(ctx context.Context, namespace, appName string) *corev1.Pod {
	podList := &corev1.PodList{}

	Eventually(func(g Gomega) {
		err := K8sClient.List(ctx, podList, client.InNamespace(namespace), client.MatchingLabels{"app": appName})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(podList.Items).NotTo(BeEmpty())
	}).WithContext(ctx).Within(60 * time.Second).ProbeEvery(1 * time.Second).Should(Succeed())

	for i := range podList.Items {
		pod := &podList.Items[i]
		if len(pod.OwnerReferences) > 0 {
			CheckPodReady(ctx, pod)
			return pod
		}
	}

	Fail("no deployment-managed Pod found")
	return nil
}

// AddEphemeralContainer attaches an ephemeral container to an existing Pod.
func AddEphemeralContainer(ctx context.Context, namespace, podName, containerName, image, targetContainer string) {
	pod, err := K8sClientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	updated := pod.DeepCopy()
	updated.Spec.EphemeralContainers = append(updated.Spec.EphemeralContainers, corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:    containerName,
			Image:   image,
			Command: []string{"sh", "-c", "sleep 300"},
		},
		TargetContainerName: targetContainer,
	})

	_, err = K8sClientset.CoreV1().Pods(namespace).UpdateEphemeralContainers(ctx, podName, updated, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}

// DeletePod deletes a Pod by name.
func DeletePod(ctx context.Context, namespace, name string) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	err := K8sClient.Delete(ctx, pod)
	if err != nil && !apierrors.IsNotFound(err) {
		Expect(err).NotTo(HaveOccurred())
	}
}

// CheckPodReady waits until a Pod is ready and has a node assigned.
func CheckPodReady(ctx context.Context, pod *corev1.Pod) {
	By(fmt.Sprintf("Checking readiness of Pod %s/%s", pod.GetNamespace(), pod.GetName()))

	Eventually(func() bool {
		if err := K8sClient.Get(ctx, client.ObjectKeyFromObject(pod), pod); err != nil {
			return false
		}
		if pod.Spec.NodeName == "" {
			return false
		}
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
				return true
			}
		}
		return false
	}, 2*time.Minute, 2*time.Second).Should(BeTrue(),
		fmt.Sprintf("Pod %s/%s did not become ready", pod.GetNamespace(), pod.GetName()))
}

// CheckResourceReadiness waits until a Deployment is ready.
func CheckResourceReadiness(ctx context.Context, resource client.Object) {
	By(fmt.Sprintf("Checking readiness of %T %s/%s", resource, resource.GetNamespace(), resource.GetName()))

	Eventually(func() bool {
		if err := K8sClient.Get(ctx, client.ObjectKeyFromObject(resource), resource); err != nil {
			return false
		}

		switch obj := resource.(type) {
		case *appsv1.Deployment:
			replicas := int32(-1)
			if obj.Spec.Replicas != nil {
				replicas = *obj.Spec.Replicas
			}
			return obj.Status.ReadyReplicas == replicas
		default:
			return false
		}
	}, 1*time.Minute, 1*time.Second).Should(BeTrue(),
		fmt.Sprintf("resource %T %s/%s did not become ready", resource, resource.GetNamespace(), resource.GetName()))
}

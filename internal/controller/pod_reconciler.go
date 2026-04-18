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

	"github.com/containeroo/kube-ephemeral-container-exporter/internal/metrics"
	"github.com/containeroo/kube-ephemeral-container-exporter/internal/predicates"
	"github.com/containeroo/kube-ephemeral-container-exporter/internal/utils"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PodReconciler reconciles Pods and exposes metrics about attached ephemeral containers.
type PodReconciler struct {
	KubeClient client.Client
	Logger     logr.Logger
	Metrics    *metrics.Registry
}

// Reconcile handles Pod changes and updates ephemeral-container metrics.
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	pod := &corev1.Pod{}
	if err := r.KubeClient.Get(ctx, req.NamespacedName, pod); err != nil {
		if apierrors.IsNotFound(err) {
			r.Logger.Info("pod not found, nothing to reconcile", "namespace", req.Namespace, "name", req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	ownerKind, ownerName := utils.ResolvePodOwner(pod)

	// Reconcile is only triggered for Pods that currently have ephemeral containers
	// or had relevant ephemeral-container related changes. Since ephemeral containers
	// are effectively append-only on a Pod, a missing spec entry is not the normal
	// cleanup path. Cleanup is handled on delete events and when old metric label
	// values must be removed before recreating the series.
	r.Metrics.UpdatePodEphemeralContainers(pod, ownerKind, ownerName)

	r.Logger.Info(
		"updated pod ephemeral-container metrics",
		"namespace", pod.Namespace,
		"name", pod.Name,
		"ownerKind", ownerKind,
		"ownerName", ownerName,
		"attachedEphemeralContainers", len(pod.Spec.EphemeralContainers),
		"statusEntries", len(pod.Status.EphemeralContainerStatuses),
	)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(predicates.EphemeralContainerChanges(r.Metrics)).
		Complete(r)
}

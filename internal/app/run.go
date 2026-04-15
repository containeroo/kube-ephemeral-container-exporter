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

package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"

	"github.com/containeroo/tinyflags"

	"github.com/containeroo/kube-ephemeral-container-exporter/internal/controller"
	"github.com/containeroo/kube-ephemeral-container-exporter/internal/flag"
	"github.com/containeroo/kube-ephemeral-container-exporter/internal/logging"
	internalmetrics "github.com/containeroo/kube-ephemeral-container-exporter/internal/metrics"
	"github.com/containeroo/kube-ephemeral-container-exporter/internal/utils"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	crmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

// Run is the main function of the application.
func Run(ctx context.Context, version string, args []string, w io.Writer) error {
	flags, err := flag.ParseArgs(args, version)
	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			fmt.Fprint(w, err.Error()) // nolint:errcheck
			return nil
		}
		return fmt.Errorf("error parsing arguments: %w", err)
	}

	logger, err := logging.InitLogging(
		flags.LogDev,
		flags.LogEncoder,
		flags.LogStacktraceLevel,
		w,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	systemLog := logging.SystemLogger(logger)
	systemLog.Info("initializing kube-ephemeral-container-exporter", "version", version)

	if len(flags.OverriddenValues) > 0 {
		systemLog.Info("CLI overrides", "overrides", flags.OverriddenValues)
	}

	tlsOpts := []func(*tls.Config){}
	if !flags.EnableHTTP2 {
		systemLog.Info("disabling HTTP/2 for compatibility")
		tlsOpts = append(tlsOpts, func(c *tls.Config) {
			c.NextProtos = []string{"http/1.1"}
		})
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	metricsServerOptions := metricsserver.Options{
		BindAddress: "0",
	}
	if flags.EnableMetrics {
		metricsServerOptions = metricsserver.Options{
			BindAddress:   flags.MetricsAddr,
			SecureServing: flags.SecureMetrics,
			TLSOpts:       tlsOpts,
		}
		if flags.SecureMetrics {
			metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
		}
	}

	cacheOpts := utils.ToCacheOptions(flags.WatchNamespaces)

	restCfg, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("unable to get Kubernetes REST config: %w", err)
	}

	mgr, err := ctrl.NewManager(restCfg, ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		Logger:                 logging.ControllerLogger(logger),
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: flags.ProbeAddr,
		LeaderElection:         flags.LeaderElection,
		LeaderElectionID:       "fc1fdccd.kube-ephemeral-container-exporter.containeroo.ch",
		Cache:                  cacheOpts,
	})
	if err != nil {
		return fmt.Errorf("unable to create manager: %w", err)
	}

	if len(flags.WatchNamespaces) == 0 {
		systemLog.Info("namespace scope", "mode", "cluster-wide")
	} else {
		systemLog.Info("namespace scope", "mode", "namespaced", "namespaces", flags.WatchNamespaces)
	}

	metricsReg := internalmetrics.NewRegistry(crmetrics.Registry)

	if err := (&controller.PodReconciler{
		Logger:     logging.ControllerLogger(logger),
		KubeClient: mgr.GetClient(),
		Metrics:    metricsReg,
		Recorder:   mgr.GetEventRecorderFor("pod-controller"),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create Pod controller: %w", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("failed to set up health check: %w", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("failed to set up ready check: %w", err)
	}

	systemLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("manager encountered an error while running: %w", err)
	}

	return nil
}

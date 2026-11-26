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

package cli

import (
	"net"

	"github.com/containeroo/tinyflags"
)

// Flags holds all configuration options for the application.
type Flags struct {
	WatchNamespaces    []string            // Namespaces to watch
	MetricsAddr        string              // Address for the metrics server
	LeaderElection     bool                // Enable leader election
	ProbeAddr          string              // Address for health and readiness probes
	SecureMetrics      bool                // Serve metrics over HTTPS
	EnableHTTP2        bool                // Enable HTTP/2 for servers
	EnableMetrics      bool                // Enable or disable metrics
	LogEncoder         string              // Log format: "json" or "console"
	LogStacktraceLevel string              // Stacktrace log level
	LogDev             bool                // Enable development logging mode
	Overrides          tinyflags.Overrides // Overridden values from environment.
}

// ParseArgs parses CLI flags into Options and handles --help/--version output.
func ParseArgs(args []string, version string) (flags Flags, err error) {
	tf := tinyflags.NewFlagSet("kube-ephemeral-container-exporter", tinyflags.ContinueOnError)
	tf.Version(version)
	tf.EnvPrefix("EC_EXPORTER_")
	tf.HideEnvs()

	// Controller
	tf.StringSliceVar(&flags.WatchNamespaces, "watch-namespace", nil, "Namespaces to watch (can be repeated or comma-separated)").
		Placeholder("NAMESPACE").
		Value()

	// Metrics
	tf.BoolVar(&flags.EnableMetrics, "metrics-enabled", true, "Enable or disable the metrics endpoint").
		Strict().
		HideAllowed().
		Value()
	metricsBindAddress := tf.TCPAddr("metrics-bind-address", &net.TCPAddr{IP: nil, Port: 8443}, "Metrics server address").
		Placeholder("ADDR:PORT").
		Value()
	tf.BoolVar(&flags.SecureMetrics, "metrics-secure", true, "Serve metrics over HTTPS").
		Strict().
		HideAllowed().
		Value()

	// Server
	healthProbeaddress := tf.TCPAddr("health-probe-bind-address", &net.TCPAddr{IP: nil, Port: 8081}, "Health and readiness probe address").
		Placeholder("ADDR:PORT").
		Value()
	tf.BoolVar(&flags.EnableHTTP2, "enable-http2", false, "Enable HTTP/2 for servers").
		Strict().
		HideAllowed().
		Value()
	tf.BoolVar(&flags.LeaderElection, "leader-elect", true, "Enable leader election").
		Strict().
		HideAllowed().
		Value()

	// Logging
	tf.StringVar(&flags.LogEncoder, "log-encoder", "json", "Log format (json, console)").
		Choices("json", "console").
		HideAllowed().
		Value()
	tf.BoolVar(&flags.LogDev, "log-devel", false, "Enable development mode logging").Value()
	tf.StringVar(&flags.LogStacktraceLevel, "log-stacktrace-level", "panic", "Stacktrace log level").
		Choices("info", "error", "panic").
		HideAllowed().
		Value()

	if err := tf.Parse(args); err != nil {
		return Flags{}, err
	}

	flags.MetricsAddr = (*metricsBindAddress).String()
	flags.ProbeAddr = (*healthProbeaddress).String()
	flags.Overrides = tf.Overrides()

	return flags, nil
}

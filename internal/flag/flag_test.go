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

package flag

import (
	"os"
	"strings"
	"testing"

	"github.com/containeroo/tinyflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "EC_EXPORTER__") {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		key := parts[0]
		value, ok := os.LookupEnv(key)
		if ok {
			t.Cleanup(func() {
				_ = os.Setenv(key, value)
			})
		} else {
			t.Cleanup(func() {
				_ = os.Unsetenv(key)
			})
		}
		_ = os.Unsetenv(key)
	}
}

func TestParseArgs(t *testing.T) {
	t.Run("use defaults", func(t *testing.T) {
		clearEnv(t)

		opts, err := ParseArgs(nil, "vX.Y.Z")
		require.NoError(t, err)

		assert.Empty(t, opts.WatchNamespaces)
		assert.Equal(t, ":8443", opts.MetricsAddr)
		assert.True(t, opts.EnableMetrics)
		assert.True(t, opts.SecureMetrics)
		assert.Equal(t, ":8081", opts.ProbeAddr)
		assert.False(t, opts.EnableHTTP2)
		assert.True(t, opts.LeaderElection)
		assert.Equal(t, "json", opts.LogEncoder)
		assert.Equal(t, "panic", opts.LogStacktraceLevel)
		assert.False(t, opts.LogDev)
		assert.Empty(t, opts.OverriddenValues)
	})

	t.Run("show version", func(t *testing.T) {
		clearEnv(t)

		_, err := ParseArgs([]string{"--version"}, "1.2.3")
		require.Error(t, err)
		assert.True(t, tinyflags.IsVersionRequested(err))
		assert.EqualError(t, err, "1.2.3")
	})

	t.Run("show help", func(t *testing.T) {
		clearEnv(t)

		_, err := ParseArgs([]string{"--help"}, "")
		require.Error(t, err)
		assert.True(t, tinyflags.IsHelpRequested(err))

		usage := err.Error()
		assert.True(t, strings.HasPrefix(usage, "Usage: kube-ephemeral-container-exporter [flags]"))
		assert.Contains(t, usage, "--watch-namespace")
		assert.Contains(t, usage, "--metrics-enabled")
		assert.Contains(t, usage, "--metrics-bind-address")
		assert.Contains(t, usage, "--metrics-secure")
		assert.Contains(t, usage, "--health-probe-bind-address")
		assert.Contains(t, usage, "--enable-http2")
		assert.Contains(t, usage, "--leader-elect")
		assert.Contains(t, usage, "--log-encoder")
		assert.Contains(t, usage, "--log-stacktrace-level")
		assert.Contains(t, usage, "--log-devel")
	})

	t.Run("custom values", func(t *testing.T) {
		clearEnv(t)

		args := []string{
			"--watch-namespace", "default,kube-system",
			"--metrics-enabled=false",
			"--metrics-bind-address", "127.0.0.1:9443",
			"--metrics-secure=false",
			"--health-probe-bind-address", "127.0.0.1:18081",
			"--enable-http2=true",
			"--leader-elect=false",
			"--log-encoder", "console",
			"--log-stacktrace-level", "error",
			"--log-devel",
		}

		opts, err := ParseArgs(args, "0.0.0")
		require.NoError(t, err)

		assert.Equal(t, []string{"default", "kube-system"}, opts.WatchNamespaces)
		assert.False(t, opts.EnableMetrics)
		assert.Equal(t, "127.0.0.1:9443", opts.MetricsAddr)
		assert.False(t, opts.SecureMetrics)
		assert.Equal(t, "127.0.0.1:18081", opts.ProbeAddr)
		assert.True(t, opts.EnableHTTP2)
		assert.False(t, opts.LeaderElection)
		assert.Equal(t, "console", opts.LogEncoder)
		assert.Equal(t, "error", opts.LogStacktraceLevel)
		assert.True(t, opts.LogDev)
	})

	t.Run("parsing error", func(t *testing.T) {
		clearEnv(t)

		_, err := ParseArgs([]string{"--invalid"}, "")
		require.Error(t, err)
		assert.EqualError(t, err, "unknown flag: --invalid")
	})

	t.Run("invalid log encoder", func(t *testing.T) {
		clearEnv(t)

		_, err := ParseArgs([]string{"--log-encoder", "xml"}, "")
		require.Error(t, err)
		assert.EqualError(t, err, "invalid value for flag --log-encoder: must be one of: json, console.")
	})

	t.Run("invalid stacktrace level", func(t *testing.T) {
		clearEnv(t)

		_, err := ParseArgs([]string{"--log-stacktrace-level", "debug"}, "")
		require.Error(t, err)
		assert.EqualError(t, err, "invalid value for flag --log-stacktrace-level: must be one of: info, error, panic.")
	})

	t.Run("multiple watch namespaces", func(t *testing.T) {
		clearEnv(t)

		args := []string{
			"--watch-namespace", "default",
			"--watch-namespace", "kube-system",
		}

		opts, err := ParseArgs(args, "")
		require.NoError(t, err)
		assert.Equal(t, []string{"default", "kube-system"}, opts.WatchNamespaces)
	})

	t.Run("env overrides", func(t *testing.T) {
		clearEnv(t)

		t.Setenv("EC_EXPORTER__LOG_ENCODER", "console")
		t.Setenv("EC_EXPORTER__METRICS_SECURE", "false")
		t.Setenv("EC_EXPORTER__LEADER_ELECT", "false")

		opts, err := ParseArgs(nil, "")
		require.NoError(t, err)

		assert.Equal(t, "console", opts.LogEncoder)
		assert.False(t, opts.SecureMetrics)
		assert.False(t, opts.LeaderElection)

		require.NotEmpty(t, opts.OverriddenValues)
		assert.Contains(t, opts.OverriddenValues, "log-encoder")
		assert.Contains(t, opts.OverriddenValues, "metrics-secure")
		assert.Contains(t, opts.OverriddenValues, "leader-elect")
	})
}

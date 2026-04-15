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
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Run("Smoke", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		args := []string{
			"--leader-elect=false",
			"--watch-namespace=test-kube-ephemeral-container-exporter",
			"--metrics-enabled=false",
			"--health-probe-bind-address=:0",
		}
		out := &bytes.Buffer{}

		errCh := make(chan error, 1)
		go func() {
			errCh <- Run(ctx, "v0.0.0", args, out)
		}()

		time.Sleep(2 * time.Second)
		cancel()

		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("Run returned an error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("Run did not return within the expected time")
		}
	})

	t.Run("Invalid args", func(t *testing.T) {
		ctx := t.Context()
		args := []string{"--invalid-flag"}
		out := &bytes.Buffer{}

		err := Run(ctx, "v0.0.0", args, out)

		require.Error(t, err)
		assert.EqualError(t, err, "error parsing arguments: unknown flag: --invalid-flag")
	})

	t.Run("Request version", func(t *testing.T) {
		ctx := t.Context()
		args := []string{"--version"}
		out := &bytes.Buffer{}

		err := Run(ctx, "v0.0.0", args, out)

		assert.NoError(t, err)
		assert.Equal(t, "v0.0.0", out.String())
	})

	t.Run("Logger error", func(t *testing.T) {
		ctx := t.Context()
		args := []string{"--log-encoder", "invalid"}
		out := &bytes.Buffer{}

		err := Run(ctx, "v0.0.0", args, out)

		require.Error(t, err)
		assert.EqualError(t, err, "error parsing arguments: invalid value for flag --log-encoder: must be one of: json, console.")
	})
}

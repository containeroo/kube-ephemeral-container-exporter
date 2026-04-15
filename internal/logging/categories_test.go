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

package logging

import (
	"bytes"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger(t *testing.T, buf *bytes.Buffer) logr.Logger {
	t.Helper()

	logger, err := InitLogging(false, EncoderConsole, LevelInfo, buf)
	require.NoError(t, err)

	return logger
}

func TestWithCategory(t *testing.T) {
	t.Parallel()

	t.Run("empty category returns logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := newTestLogger(t, &buf)
		out := WithCategory(logger, "")

		out.Info("hello")
		assert.NotContains(t, buf.String(), "category")
	})

	t.Run("category added", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := newTestLogger(t, &buf)
		out := WithCategory(logger, CategoryController)

		out.Info("hello")
		assert.Contains(t, buf.String(), "category")
		assert.Contains(t, buf.String(), "controller")
	})
}

func TestCategoryLoggers(t *testing.T) {
	t.Parallel()

	t.Run("system logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := newTestLogger(t, &buf)

		SystemLogger(logger).Info("system")

		assert.Contains(t, buf.String(), "category")
		assert.Contains(t, buf.String(), "system")
	})

	t.Run("controller logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := newTestLogger(t, &buf)

		ControllerLogger(logger).Info("controller")

		assert.Contains(t, buf.String(), "category")
		assert.Contains(t, buf.String(), "controller")
	})
}

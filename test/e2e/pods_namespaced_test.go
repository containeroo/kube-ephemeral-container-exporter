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

package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/containeroo/kube-ephemeral-container-exporter/test/testutils"

	. "github.com/onsi/ginkgo/v2" // nolint:staticcheck
)

var _ = Describe("Namespaced mode", Serial, Ordered, func() {
	var watchedNS string

	BeforeAll(func() {
		By("stopping any running operator instance")
		testutils.StopOperator()
		time.Sleep(4 * time.Second)

		By("resetting log buffer before suite")
		testutils.LogBuffer.Reset()

		By("creating the watched namespace")
		watchedNS = testutils.NSManager.CreateNamespace(context.Background())

		By("starting operator in namespaced mode")
		testutils.StartOperatorWithFlags([]string{
			"--leader-elect=false",
			"--metrics-enabled=true",
			"--metrics-secure=false",
			"--metrics-bind-address=127.0.0.1:18080",
			"--health-probe-bind-address=127.0.0.1:18081",
			"--watch-namespace=" + watchedNS,
		})
	})

	AfterAll(func(ctx SpecContext) {
		By("stopping operator after suite")
		testutils.StopOperator()

		By("cleaning up watched namespace")
		testutils.NSManager.Cleanup(ctx)
	})

	It("exports metrics for pods within the watched namespace", func(ctx SpecContext) {
		name := testutils.GenerateUniqueName("pod")

		By("creating a pod in the watched namespace")
		pod := testutils.CreatePod(ctx, watchedNS, name)

		By("adding an ephemeral container")
		testutils.AddEphemeralContainer(ctx, pod.Namespace, pod.Name, "debugger", "busybox:1.36", testutils.DefaultTestImageName)

		By("waiting for metrics to appear")
		testutils.MetricsContain(
			fmt.Sprintf(`namespace=%q`, pod.Namespace),
			90*time.Second,
			2*time.Second,
		)
		testutils.MetricsContain(
			fmt.Sprintf(`pod=%q`, pod.Name),
			90*time.Second,
			2*time.Second,
		)
	})

	It("does not export metrics for pods outside the watched namespace", func(ctx SpecContext) {
		By("creating a separate namespace not watched by the operator")
		otherNS := testutils.NSManager.CreateNamespace(ctx)

		name := testutils.GenerateUniqueName("pod")

		By("creating a pod outside the watched namespace")
		pod := testutils.CreatePod(ctx, otherNS, name)

		By("adding an ephemeral container")
		testutils.AddEphemeralContainer(ctx, pod.Namespace, pod.Name, "debugger", "busybox:1.36", testutils.DefaultTestImageName)

		By("ensuring no metrics are exported for the pod outside the watched namespace")
		testutils.MetricsConsistentlyNotContain(
			fmt.Sprintf(`pod=%q`, pod.Name),
			12*time.Second,
			1*time.Second,
		)
	})
})

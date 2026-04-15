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
	"fmt"
	"time"

	"github.com/containeroo/kube-ephemeral-container-exporter/test/testutils"

	. "github.com/onsi/ginkgo/v2" // nolint:staticcheck
	. "github.com/onsi/gomega"    // nolint:staticcheck
)

var _ = Describe("Pods generic", Serial, Ordered, func() {
	var ns string

	BeforeAll(func() {
		By("stopping any running operator instance")
		testutils.StopOperator()
		time.Sleep(4 * time.Second)

		By("resetting log buffer before suite")
		testutils.LogBuffer.Reset()

		By("starting operator with metrics enabled")
		testutils.StartOperatorWithFlags([]string{
			"--leader-elect=false",
			"--metrics-enabled=true",
			"--metrics-secure=false",
			"--metrics-bind-address=127.0.0.1:18080",
			"--health-probe-bind-address=127.0.0.1:18081",
		})
	})

	AfterAll(func() {
		By("stopping operator after suite")
		testutils.StopOperator()
	})

	BeforeEach(func(ctx SpecContext) {
		By("creating a fresh namespace for the test")
		ns = testutils.NSManager.CreateNamespace(ctx)
	})

	AfterEach(func(ctx SpecContext) {
		By("cleaning up namespace and resetting logs")
		testutils.NSManager.Cleanup(ctx)
		testutils.LogBuffer.Reset()
	})

	It("exports metrics for a pod with an ephemeral container", func(ctx SpecContext) {
		name := testutils.GenerateUniqueName("pod")

		By("creating a pod")
		pod := testutils.CreatePod(ctx, ns, name)

		By("ensuring no ephemeral metrics exist for the pod yet")
		testutils.MetricsConsistentlyNotContain(
			fmt.Sprintf(`pod=%q`, pod.Name),
			4*time.Second,
			1*time.Second,
		)

		By("adding an ephemeral container")
		testutils.AddEphemeralContainer(ctx, pod.Namespace, pod.Name, "debugger", "busybox:1.36", testutils.DefaultTestImageName)

		By("waiting for pod-level and container-level metrics to be exported")
		Eventually(func(g Gomega) {
			refreshedPod, err := testutils.GetPod(ctx, pod.Namespace, pod.Name)
			g.Expect(err).NotTo(HaveOccurred())

			metricsText, err := testutils.ScrapeMetrics()
			g.Expect(err).NotTo(HaveOccurred())

			// Label order in Prometheus exposition is not guaranteed, so we assert
			// stable fragments instead of one exact full sample line.
			g.Expect(metricsText).To(ContainSubstring(`kube_pod_ephemeral_container_present{`))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`namespace=%q`, refreshedPod.Namespace)))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`pod=%q`, refreshedPod.Name)))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`node=%q`, refreshedPod.Spec.NodeName)))
			g.Expect(metricsText).To(ContainSubstring(`owner_kind=""`))
			g.Expect(metricsText).To(ContainSubstring(`owner_name=""`))
			g.Expect(metricsText).To(ContainSubstring(`} 1`))

			g.Expect(metricsText).To(ContainSubstring(`kube_pod_ephemeral_container_count{`))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`namespace=%q`, refreshedPod.Namespace)))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`pod=%q`, refreshedPod.Name)))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`node=%q`, refreshedPod.Spec.NodeName)))
			g.Expect(metricsText).To(ContainSubstring(`owner_kind=""`))
			g.Expect(metricsText).To(ContainSubstring(`owner_name=""`))
			g.Expect(metricsText).To(ContainSubstring(`} 1`))

			g.Expect(metricsText).To(ContainSubstring(`kube_pod_ephemeral_container_info{`))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`namespace=%q`, refreshedPod.Namespace)))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`pod=%q`, refreshedPod.Name)))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`node=%q`, refreshedPod.Spec.NodeName)))
			g.Expect(metricsText).To(ContainSubstring(`owner_kind=""`))
			g.Expect(metricsText).To(ContainSubstring(`owner_name=""`))
			g.Expect(metricsText).To(ContainSubstring(`container="debugger"`))
			g.Expect(metricsText).To(ContainSubstring(`image="busybox:1.36"`))
			g.Expect(metricsText).To(ContainSubstring(`} 1`))

			g.Expect(metricsText).To(ContainSubstring(`kube_pod_ephemeral_container_running{`))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`namespace=%q`, refreshedPod.Namespace)))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`pod=%q`, refreshedPod.Name)))
			g.Expect(metricsText).To(ContainSubstring(`container="debugger"`))
			g.Expect(metricsText).To(ContainSubstring(`} 1`))
		}).WithContext(ctx).Within(90 * time.Second).ProbeEvery(2 * time.Second).Should(Succeed())

		By("verifying the reconcile log was emitted")
		testutils.ContainsLogs(
			fmt.Sprintf(`"name":%q`, pod.Name),
			10*time.Second,
			1*time.Second,
		)
	})

	It("removes metrics when a pod is deleted", func(ctx SpecContext) {
		name := testutils.GenerateUniqueName("pod")

		By("creating a pod")
		pod := testutils.CreatePod(ctx, ns, name)

		By("adding an ephemeral container")
		testutils.AddEphemeralContainer(ctx, pod.Namespace, pod.Name, "debugger", "busybox:1.36", testutils.DefaultTestImageName)

		By("waiting for metrics to appear")
		testutils.MetricsContain(
			fmt.Sprintf(`pod=%q`, pod.Name),
			90*time.Second,
			2*time.Second,
		)

		By("deleting the pod")
		testutils.DeletePod(ctx, pod.Namespace, pod.Name)

		By("waiting for metrics to be removed")
		Eventually(func(g Gomega) {
			metricsText, err := testutils.ScrapeMetrics()
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(metricsText).NotTo(ContainSubstring(fmt.Sprintf(`pod=%q`, pod.Name)))
		}).WithContext(ctx).Within(90 * time.Second).ProbeEvery(2 * time.Second).Should(Succeed())
	})

	It("exports owner labels for a deployment-managed pod", func(ctx SpecContext) {
		name := testutils.GenerateUniqueName("dep")

		By("creating a deployment")
		dep := testutils.CreateDeployment(ctx, ns, name)

		By("finding the created pod")
		pod := testutils.GetDeploymentPod(ctx, dep.Namespace, dep.Name)

		By("adding an ephemeral container to the deployment pod")
		testutils.AddEphemeralContainer(ctx, pod.Namespace, pod.Name, "debugger", "busybox:1.36", testutils.DefaultTestImageName)

		By("waiting for owner labels to appear in metrics")
		Eventually(func(g Gomega) {
			refreshedPod, err := testutils.GetPod(ctx, pod.Namespace, pod.Name)
			g.Expect(err).NotTo(HaveOccurred())

			ownerKind, ownerName := testutils.GetDirectOwner(refreshedPod)
			g.Expect(ownerKind).To(Equal("ReplicaSet"))
			g.Expect(ownerName).NotTo(BeEmpty())

			metricsText, err := testutils.ScrapeMetrics()
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(metricsText).To(ContainSubstring(`kube_pod_ephemeral_container_present{`))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`namespace=%q`, refreshedPod.Namespace)))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`pod=%q`, refreshedPod.Name)))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`node=%q`, refreshedPod.Spec.NodeName)))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`owner_kind=%q`, ownerKind)))
			g.Expect(metricsText).To(ContainSubstring(fmt.Sprintf(`owner_name=%q`, ownerName)))
			g.Expect(metricsText).To(ContainSubstring(`} 1`))
		}).WithContext(ctx).Within(90 * time.Second).ProbeEvery(2 * time.Second).Should(Succeed())
	})
})

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
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/gomega" // nolint:staticcheck
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sClient is the shared Kubernetes client used in e2e tests.
var K8sClient client.Client

// K8sClientset is the shared typed Kubernetes client used in e2e tests.
var K8sClientset kubernetes.Interface

// GenerateUniqueName returns a unique name based on the given base string and a truncated UUID.
func GenerateUniqueName(base string) string {
	return fmt.Sprintf("%s-%s", base, uuid.New().String()[:5])
}

// Int32Ptr returns a pointer to the given int32 value.
func Int32Ptr(i int32) *int32 {
	return &i
}

// ScrapeMetrics fetches the exporter metrics endpoint.
func ScrapeMetrics() (string, error) {
	resp, err := http.Get("http://127.0.0.1:18080/metrics") // nolint:gosec
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// MetricsContain checks if the expected metric text is present.
func MetricsContain(expected string, timeout, interval time.Duration) {
	Eventually(func() bool {
		metricsText, err := ScrapeMetrics()
		if err != nil {
			return false
		}
		return strings.Contains(metricsText, expected)
	}, timeout, interval).Should(BeTrue(), fmt.Sprintf("Expected metric text not found: %s", expected))
}

// MetricsConsistentlyNotContain checks that the expected metric text is absent
// for the full duration.
func MetricsConsistentlyNotContain(expected string, timeout, interval time.Duration) {
	Consistently(func() bool {
		metricsText, err := ScrapeMetrics()
		if err != nil {
			return false
		}
		return strings.Contains(metricsText, expected)
	}, timeout, interval).Should(BeFalse(), fmt.Sprintf("Expected metric text should not be found: %s", expected))
}

// MetricsEventuallyNotContain checks that the expected metric text eventually disappears.
func MetricsEventuallyNotContain(expected string, timeout, interval time.Duration) {
	Eventually(func() bool {
		metricsText, err := ScrapeMetrics()
		if err != nil {
			return false
		}
		return !strings.Contains(metricsText, expected)
	}, timeout, interval).Should(BeTrue(), fmt.Sprintf("Expected metric text should eventually disappear: %s", expected))
}

// GetDirectOwner returns the direct controller owner of the Pod if one exists.
func GetDirectOwner(pod interface {
	GetOwnerReferences() []metav1.OwnerReference
},
) (string, string) {
	for _, ref := range pod.GetOwnerReferences() {
		if ref.Controller != nil && *ref.Controller {
			return ref.Kind, ref.Name
		}
	}
	return "", ""
}

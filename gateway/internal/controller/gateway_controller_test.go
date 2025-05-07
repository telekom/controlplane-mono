/*
Copyright 2024.

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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/telekom/controlplane-mono/common/pkg/condition"
	"github.com/telekom/controlplane-mono/common/pkg/config"
	gatewayv1 "github.com/telekom/controlplane-mono/gateway/api/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewGateway(name string) *gatewayv1.Gateway {
	return &gatewayv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
			Labels: map[string]string{
				config.EnvironmentLabelKey:   testEnvironment,
				config.BuildLabelKey("zone"): "test",
			},
		},
		Spec: gatewayv1.GatewaySpec{
			Admin: gatewayv1.AdminConfig{
				ClientId:     "admin",
				ClientSecret: "topsecret",
				IssuerUrl:    "https://issuer.url",
				Url:          fmt.Sprintf("https://admin.%s.url", name),
			},
		},
	}
}

var _ = Describe("Gateway Controller", Ordered, func() {

	var gateway *gatewayv1.Gateway

	BeforeAll(func() {
		By("Initializing the Gateway")
		gateway = NewGateway("test-gateway")
	})

	AfterAll(func() {
		By("Tearing down the Gateway")
		err := k8sClient.Delete(ctx, gateway)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Creating a Gateway", func() {
		It("should be ready", func() {
			err := k8sClient.Create(ctx, gateway)
			Expect(err).NotTo(HaveOccurred())

			By("Checking if the Gateway is ready")
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(gateway), gateway)
				g.Expect(err).NotTo(HaveOccurred())
				By("Checking the conditions")
				g.Expect(gateway.Status.Conditions).To(HaveLen(2))
				readyCondition := meta.FindStatusCondition(gateway.Status.Conditions, condition.ConditionTypeReady)
				g.Expect(readyCondition).NotTo(BeNil())
				g.Expect(readyCondition.Status).To(Equal(metav1.ConditionTrue))

			}, timeout, interval).Should(Succeed())
		})
	})
})

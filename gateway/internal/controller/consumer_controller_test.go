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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/telekom/controlplane-mono/common/pkg/condition"
	"github.com/telekom/controlplane-mono/common/pkg/config"
	"github.com/telekom/controlplane-mono/common/pkg/types"
	gatewayv1 "github.com/telekom/controlplane-mono/gateway/api/v1"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewConsumer(name string, realmRef types.ObjectRef) *gatewayv1.Consumer {
	return &gatewayv1.Consumer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
			Labels: map[string]string{
				config.EnvironmentLabelKey:    testEnvironment,
				config.BuildLabelKey("realm"): realmRef.Name,
			},
		},
		Spec: gatewayv1.ConsumerSpec{
			Realm: realmRef,
			Name:  name,
		},
	}
}

var _ = Describe("Consumer Controller", Ordered, func() {

	var gateway *gatewayv1.Gateway
	var realm *gatewayv1.Realm

	var consumer *gatewayv1.Consumer

	BeforeAll(func() {
		By("Creating the Gateway and Realm")
		gateway = NewGateway("test-consumer")
		err := k8sClient.Create(ctx, gateway)
		Expect(err).NotTo(HaveOccurred())
		realm = NewRealm("test-consumer")
		realm.Spec.Gateway = types.ObjectRefFromObject(gateway)

		err = k8sClient.Create(ctx, realm)
		Expect(err).NotTo(HaveOccurred())

		By("Initializing the Consumer")
		consumer = NewConsumer("test-consumer", *types.ObjectRefFromObject(realm))

	})

	AfterAll(func() {
		By("Cleaning up the resources")
		err := k8sClient.Delete(ctx, gateway)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(ctx, realm)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("When creating a Consumer", func() {

		It("should successfully provision the Consumer", func() {
			By("Creating the Consumer")
			err := k8sClient.Create(ctx, consumer)
			Expect(err).NotTo(HaveOccurred())

			By("Setting up the mocks")
			GetMockClientFor(gateway).EXPECT().CreateOrReplaceConsumer(gomock.Any(), consumer.Name).Return(nil).MinTimes(1)

			By("Checking the status")
			Eventually(func(g Gomega) {
				By("fetching the realm")
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(realm), realm)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(meta.IsStatusConditionTrue(realm.GetConditions(), condition.ConditionTypeReady)).To(BeTrue())

				By("fetching the consumer")
				err = k8sClient.Get(ctx, client.ObjectKeyFromObject(consumer), consumer)
				g.Expect(err).NotTo(HaveOccurred())

				By("checking the conditions")
				g.Expect(consumer.GetConditions()).ToNot(BeEmpty())
				readyCondition := meta.FindStatusCondition(consumer.GetConditions(), condition.ConditionTypeReady)
				g.Expect(readyCondition).NotTo(BeNil())
				g.Expect(readyCondition.Reason).To(Equal("ConsumerReady"))
				g.Expect(readyCondition.Status).To(Equal(metav1.ConditionTrue))

			}, 5*time.Second, interval).Should(Succeed())

		})

		It("should delete the Consumer", func() {
			By("Deleting the Consumer")
			err := k8sClient.Delete(ctx, consumer)
			Expect(err).NotTo(HaveOccurred())

			By("Setting up the mocks")
			GetMockClientFor(gateway).EXPECT().DeleteConsumer(gomock.Any(), consumer.Spec.Name).Return(nil).MinTimes(1)

			By("Checking the status")
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(consumer), consumer)
				g.Expect(err).To(HaveOccurred())

			}, timeout, interval).Should(Succeed())
		})
	})
})

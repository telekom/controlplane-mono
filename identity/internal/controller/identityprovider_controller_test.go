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
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/telekom/controlplane-mono/common/pkg/condition"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"

	identityv1 "github.com/telekom/controlplane-mono/identity/api/v1"
	identityproviderModel "github.com/telekom/controlplane-mono/identity/internal/model/identityprovider"
)

var _ = Describe("IdentityProvider Controller", func() {
	Context("When reconciling a resource", func() {
		ctx := context.Background()

		// IDP related
		identityProviderName := "keycloak-idp"
		idpRef := client.ObjectKey{
			Name:      identityProviderName,
			Namespace: testNamespace,
		}
		testIdp := identityproviderModel.NewIdentityProvider(identityProviderName, testNamespace, testEnvironment)

		BeforeEach(func() {
			By("creating the custom resource for the Kind IdentityProvider")
			NewIdentityProvider(ctx, idpRef, testIdp)
		})

		AfterEach(func() {
			By("deleting the custom resource for the Kind IdentityProvider")
			DeleteIdentityProvider(ctx, idpRef)
		})
		It("should successfully reconcile the resource", func() {
			Eventually(func(g Gomega) {
				VerifyIdentityProvider(ctx, g, idpRef, testIdp)

			}, timeout, interval).Should(Succeed())
		})
	})
})

func VerifyIdentityProvider(ctx context.Context, gomega Gomega, namespacedName client.ObjectKey, idpToVerify *identityv1.IdentityProvider) {
	idpResource := &identityv1.IdentityProvider{}
	err := k8sClient.Get(ctx, namespacedName, idpResource)

	gomega.Expect(err).NotTo(HaveOccurred())

	gomega.Expect(idpResource.Spec).To(Equal(idpToVerify.Spec))
	gomega.Expect(idpResource.Status).NotTo(BeNil())
	gomega.Expect(idpResource.Status.AdminUrl).To(Equal(idpToVerify.Spec.AdminUrl))
	gomega.Expect(idpResource.Status.AdminTokenUrl).NotTo(BeEmpty())
	gomega.Expect(idpResource.Status.AdminConsoleUrl).NotTo(BeEmpty())
	gomega.Expect(idpResource.Status.Conditions).To(HaveLen(2))
	gomega.Expect(meta.IsStatusConditionTrue(idpResource.Status.Conditions, condition.ConditionTypeProcessing)).To(BeFalse())
	gomega.Expect(meta.IsStatusConditionTrue(idpResource.Status.Conditions, condition.ConditionTypeReady)).To(BeTrue())
}

func NewIdentityProvider(ctx context.Context, namespacedName client.ObjectKey, idp *identityv1.IdentityProvider) {
	idpResource := &identityv1.IdentityProvider{}
	err := k8sClient.Get(ctx, namespacedName, idpResource)
	if err != nil && errors.IsNotFound(err) {
		Expect(k8sClient.Create(ctx, idp)).To(Succeed())
	}
}

func DeleteIdentityProvider(ctx context.Context, namespacedName client.ObjectKey) {
	idpResource := &identityv1.IdentityProvider{}
	err := k8sClient.Get(ctx, namespacedName, idpResource)
	Expect(err).NotTo(HaveOccurred())

	Expect(k8sClient.Delete(ctx, idpResource)).To(Succeed())
}

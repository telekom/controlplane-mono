package controller_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/telekom/controlplane-mono/secret-manager/pkg/backend"
	"github.com/telekom/controlplane-mono/secret-manager/pkg/controller"
	"github.com/telekom/controlplane-mono/secret-manager/test/mocks"
)

var _ = Describe("Controller", func() {
	BeforeEach(func() {

	})

	Context("Secrets Controller", func() {

		var mockedBackend *mocks.MockBackend[*mocks.MockSecretId, backend.DefaultSecret[*mocks.MockSecretId]]

		BeforeEach(func() {
			mockedBackend = mocks.NewMockBackend[*mocks.MockSecretId, backend.DefaultSecret[*mocks.MockSecretId]](GinkgoT())
		})

		It("should get a secret", func() {
			ctx := context.Background()
			ctrl := controller.NewSecretsController(mockedBackend)

			secretId := mocks.NewMockSecretId(GinkgoT())
			secretId.EXPECT().String().Return("my-secret-id")

			mockedBackend.EXPECT().ParseSecretId("my-secret-id").Return(secretId, nil).Times(1)
			mockedBackend.EXPECT().Get(ctx, secretId).Return(backend.NewDefaultSecret(secretId, "my-value"), nil).Times(1)

			res, err := ctrl.GetSecret(ctx, "my-secret-id")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).ToNot(BeNil())
			Expect(res.Id).To(Equal("my-secret-id"))
			Expect(res.Value).To(Equal("my-value"))
		})

		It("should set a secret", func() {
			ctx := context.Background()
			ctrl := controller.NewSecretsController(mockedBackend)

			secretId := mocks.NewMockSecretId(GinkgoT())
			secretId.EXPECT().String().Return("my-secret-id")

			mockedBackend.EXPECT().ParseSecretId("my-secret-id").Return(secretId, nil).Times(1)
			mockedBackend.EXPECT().Set(ctx, secretId, backend.String("my-value")).Return(backend.NewDefaultSecret(secretId, "my-value"), nil).Times(1)

			res, err := ctrl.SetSecret(ctx, "my-secret-id", "my-value")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).ToNot(BeNil())
			Expect(res.Id).To(Equal("my-secret-id"))
			Expect(res.Value).To(Equal("my-value"))
		})

		It("should delete a secret", func() {
			ctx := context.Background()
			ctrl := controller.NewSecretsController(mockedBackend)

			secretId := mocks.NewMockSecretId(GinkgoT())
			secretId.EXPECT().String().Return("my-secret-id")

			mockedBackend.EXPECT().ParseSecretId("my-secret-id").Return(secretId, nil).Times(1)
			mockedBackend.EXPECT().Delete(ctx, secretId).Return(nil).Times(1)

			err := ctrl.DeleteSecret(ctx, "my-secret-id")
			Expect(err).ToNot(HaveOccurred())
		})

	})

	Context("Onboard Controller", func() {

		var mockedOnboarder *mocks.MockOnboarder

		BeforeEach(func() {
			mockedOnboarder = mocks.NewMockOnboarder(GinkgoT())
		})

		It("should onboard an environment", func() {
			ctx := context.Background()
			ctrl := controller.NewOnboardController(mockedOnboarder)

			mockedOnboarder.EXPECT().OnboardEnvironment(ctx, "env-id").Return(backend.NewDefaultOnboardResponse(nil), nil).Times(1)

			res, err := ctrl.OnboardEnvironment(ctx, "env-id")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).ToNot(BeNil())
		})

		It("should delete an environment", func() {
			ctx := context.Background()
			ctrl := controller.NewOnboardController(mockedOnboarder)

			mockedOnboarder.EXPECT().DeleteEnvironment(ctx, "env-id").Return(nil).Times(1)

			err := ctrl.DeleteEnvironment(ctx, "env-id")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should onboard a team", func() {
			ctx := context.Background()
			ctrl := controller.NewOnboardController(mockedOnboarder)

			mockedOnboarder.EXPECT().OnboardTeam(ctx, "env-id", "team-id").Return(backend.NewDefaultOnboardResponse(nil), nil).Times(1)

			res, err := ctrl.OnboardTeam(ctx, "env-id", "team-id")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).ToNot(BeNil())
		})

		It("should delete a team", func() {
			ctx := context.Background()
			ctrl := controller.NewOnboardController(mockedOnboarder)

			mockedOnboarder.EXPECT().DeleteTeam(ctx, "env-id", "team-id").Return(nil).Times(1)

			err := ctrl.DeleteTeam(ctx, "env-id", "team-id")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should onboard an application", func() {
			ctx := context.Background()
			ctrl := controller.NewOnboardController(mockedOnboarder)

			mockedOnboarder.EXPECT().OnboardApplication(ctx, "env-id", "team-id", "app-id").Return(backend.NewDefaultOnboardResponse(nil), nil).Times(1)

			res, err := ctrl.OnboardApplication(ctx, "env-id", "team-id", "app-id")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).ToNot(BeNil())
		})

		It("should delete an application", func() {
			ctx := context.Background()
			ctrl := controller.NewOnboardController(mockedOnboarder)

			mockedOnboarder.EXPECT().DeleteApplication(ctx, "env-id", "team-id", "app-id").Return(nil).Times(1)

			err := ctrl.DeleteApplication(ctx, "env-id", "team-id", "app-id")
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

package conjur_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/telekom/controlplane-mono/secret-manager/pkg/backend/conjur"
)

var _ = Describe("Bouncer Test", func() {
	BeforeEach(func() {

	})

	Context("Runnables", func() {
		It("should run the function and return when its done", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			bouncer := conjur.NewBouncer(5, time.Second)
			bouncer.StartN(ctx, 1)

			result := "initial"
			runnable := func(ctx context.Context) error {
				time.Sleep(500 * time.Millisecond)
				result = "done"
				return nil
			}

			err := <-bouncer.Run(ctx, runnable)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("done"))
		})

		It("should handle errors in the runnable", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			bouncer := conjur.NewBouncer(5, time.Second)
			bouncer.StartN(ctx, 1)

			result := "initial"
			runnable := func(ctx context.Context) error {
				return fmt.Errorf("error")
			}

			err := <-bouncer.Run(ctx, runnable)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(result).To(Equal("initial"))
		})

		It("should handle a full queue", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			bouncer := conjur.NewBouncer(1, time.Second)
			// bouncer.StartN(ctx, 1) Disabled to test a full queue

			result := "initial"
			runnable := func(ctx context.Context) error {
				result = "done"
				return nil
			}

			bouncer.Run(ctx, runnable)
			err := <-bouncer.Run(ctx, runnable)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("queue is full"))
			Expect(errors.Is(err, conjur.ErrQueueFull)).To(BeTrue())

			Expect(result).To(Equal("initial"))
		})
	})
})

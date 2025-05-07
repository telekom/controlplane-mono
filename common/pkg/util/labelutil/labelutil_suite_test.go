package labelutil

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLabelutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Labelutil Suite")
}

var _ = Describe("Labelutil", func() {

	Context("NormalizeValue", func() {

		It("should normalize value", func() {

			value := " foo/bar_baz/ "
			Expect(NormalizeValue(value)).To(Equal("foo-bar-baz"))

			value = "/foo/bar_baz\\"
			Expect(NormalizeValue(value)).To(Equal("foo-bar-baz"))
		})

	})
})

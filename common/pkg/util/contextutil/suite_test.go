package contextutil

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestContextutil(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Contextutil Suite")
}

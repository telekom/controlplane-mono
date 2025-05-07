package conjur_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/telekom/controlplane-mono/secret-manager/pkg/backend/conjur"
)

func TestConjur(t *testing.T) {
	conjur.RootPolicyPath = "controlplane"

	RegisterFailHandler(Fail)
	RunSpecs(t, "Conjur Suite")
}

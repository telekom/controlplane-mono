package testutil

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Module", func() {

	Context("FindNextGoModPath", func() {
		It("should find the go.mod file", func() {
			path, err := findNextFileMatch("go.mod")
			Expect(err).To(BeNil())
			Expect(path).To(Equal("../../../../go.mod"))

			_, err = os.Stat(path)
			Expect(err).To(BeNil())
		})
	})

	Context("GetCrdPaths", func() {
		It("should load the CRD paths", func() {
			dummyModPathRE := "k8s.io/client-go"

			paths, err := GetCrdPaths(dummyModPathRE)
			Expect(err).To(BeNil())
			Expect(paths).To(HaveLen(1))
			expectedPattern := filepath.Join(goPathOrDefault(), "pkg/mod/"+dummyModPathRE) + `@v\d{1,2}\.\d{1,2}\.\d{1,2}` + "/crds"
			Expect(paths[0]).To(MatchRegexp(expectedPattern))
		})
	})

	Context("GetCrdPathsOrDie", func() {
		It("should load the CRD paths", func() {
			dummyModPathRE := "k8s.io/client-go"

			paths := GetCrdPathsOrDie(dummyModPathRE)
			Expect(paths).To(HaveLen(1))
			expectedPattern := filepath.Join(goPathOrDefault(), "pkg/mod/"+dummyModPathRE) + `@v\d{1,2}\.\d{1,2}\.\d{1,2}` + "/crds"
			Expect(paths[0]).To(MatchRegexp(expectedPattern))
		})
	})

})

package inmemory

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/telekom/controlplane-mono/common-server/pkg/problems"
	"github.com/telekom/controlplane-mono/common-server/pkg/store"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerateUnstructured(n int) []*unstructured.Unstructured {
	objs := make([]*unstructured.Unstructured, n)
	for i := 0; i < n; i++ {
		objs[i] = NewUnstructured(fmt.Sprintf("foo%d", i))
		objs[i].SetLabels(map[string]string{
			"app": fmt.Sprintf("app%d", i),
		})
	}
	return objs
}

func NewUnstructured(name string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{
		Object: map[string]any{
			"metadata": map[string]any{
				"name":      name,
				"namespace": "default",
			},
			"spec": map[string]any{
				"replicas": rand.Int64N(100), // #nosec G404 -- This is not a cryptographic use case
				"timeout":  rand.Float64(),   // #nosec G404 -- This is not a cryptographic use case
			},
		},
	}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "test",
		Version: "v1",
		Kind:    "TestObject",
	})
	return u
}

var _ = Describe("Inmemory ObjectStore", func() {

	ctx := context.Background()
	scheme := runtime.NewScheme()

	gvr := schema.GroupVersionResource{
		Group:    "test",
		Version:  "v1",
		Resource: "testobjects",
	}
	gvk := schema.GroupVersionKind{
		Group:   "test",
		Version: "v1",
		Kind:    "TestObject",
	}

	Context("Basic Store Functionalities", func() {

		fakeClient := fake.NewSimpleDynamicClient(scheme, NewUnstructured("foo"))
		objStore := NewOrDie[*unstructured.Unstructured](ctx, StoreOpts{
			Client:       fakeClient,
			GVR:          gvr,
			GVK:          gvk,
			AllowedSorts: nil,
		})

		BeforeEach(func() {
			fakeClient.ClearActions()
		})

		It("should return info", func() {
			actualGVR, actualGVK := objStore.Info()
			Expect(actualGVR).To(Equal(gvr))
			Expect(actualGVK).To(Equal(gvk))
		})

		It("should create an object", func() {
			obj := NewUnstructured("bar")
			err := objStore.CreateOrReplace(ctx, obj)
			Expect(err).ToNot(HaveOccurred())

			actions := fakeClient.Actions()
			Expect(actions).To(HaveLen(1))
			Expect(actions[0].GetVerb()).To(Equal("create"))
		})

		It("should replace an object", func() {
			obj := NewUnstructured("bar")
			err := objStore.CreateOrReplace(ctx, obj)
			Expect(err).ToNot(HaveOccurred())

			actions := fakeClient.Actions()
			Expect(actions).To(HaveLen(1))
			Expect(actions[0].GetVerb()).To(Equal("update"))
		})

		It("should get an object", func() {
			obj, err := objStore.Get(ctx, "default", "foo")
			Expect(err).ToNot(HaveOccurred())
			Expect(obj.GetName()).To(Equal("foo"))
		})

		It("should list objects", func() {
			objs, err := objStore.List(ctx, store.NewListOpts())
			Expect(err).ToNot(HaveOccurred())
			Expect(objs.Items).To(HaveLen(2))
			Expect(objs.Items[0].GetName()).To(Equal("bar"))
			Expect(objs.Items[1].GetName()).To(Equal("foo"))
		})

		It("should patch an object", func() {
			patches := []store.Patch{
				{
					Path:  "spec.replicas",
					Op:    store.OpReplace,
					Value: 100,
				},
			}
			obj, err := objStore.Patch(ctx, "default", "foo", patches...)
			Expect(err).ToNot(HaveOccurred())
			Expect(obj.GetNamespace()).To(Equal("default"))
			Expect(obj.GetName()).To(Equal("foo"))
			Expect(obj.Object["spec"].(map[string]any)["replicas"]).To(Equal(int64(100)))

			Expect(fakeClient.Actions()).To(HaveLen(1))
			Expect(fakeClient.Actions()[0].GetVerb()).To(Equal("update"))
		})

		It("should delete an object", func() {
			err := objStore.Delete(ctx, "default", "foo")
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeClient.Actions()).To(HaveLen(1))
			Expect(fakeClient.Actions()[0].GetVerb()).To(Equal("delete"))
		})

		It("should return error on delete (not found)", func() {
			err := objStore.Delete(ctx, "default", "noexist")
			Expect(err).To(HaveOccurred())
			Expect(problems.IsNotFound(err)).To(BeTrue())
		})

		It("should return error on get (not found)", func() {
			obj, err := objStore.Get(ctx, "default", "noexist")
			Expect(err).To(HaveOccurred())
			Expect(problems.IsNotFound(err)).To(BeTrue())
			Expect(obj).To(BeNil())
		})

	})

	Context("List", Ordered, func() {
		fakeClient := fake.NewSimpleDynamicClient(scheme, NewUnstructured("foo"))
		objStore := NewOrDie[*unstructured.Unstructured](ctx, StoreOpts{
			Client:       fakeClient,
			GVR:          gvr,
			GVK:          gvk,
			AllowedSorts: nil,
		})

		BeforeEach(func() {
			fakeClient.ClearActions()
		})

		BeforeAll(func() {
			for _, u := range GenerateUnstructured(500) {
				Expect(objStore.(*InmemoryObjectStore[*unstructured.Unstructured]).OnCreate(ctx, u)).To(Succeed())
			}
		})

		It("should filter by label", func() {
			listOpts := store.NewListOpts()
			listOpts.Filters = []store.Filter{
				{
					Path:  "metadata.labels.app",
					Op:    store.OpRegex,
					Value: "^app1[0-9]{1}$",
				},
			}

			list, err := objStore.List(ctx, listOpts)
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(10))
		})

		It("should filter by prefix", func() {
			listOpts := store.NewListOpts()
			listOpts.Prefix = "default/foo9"

			list, err := objStore.List(ctx, listOpts)
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(11)) // foo90 to foo99 and foo9
		})

		It("should filter by prefix and label", func() {
			listOpts := store.NewListOpts()
			listOpts.Prefix = "default/foo9"
			listOpts.Filters = []store.Filter{
				{
					Path:  "metadata.labels.app",
					Op:    store.OpRegex,
					Value: "^app9[0-9]{1}$",
				},
			}

			list, err := objStore.List(ctx, listOpts)
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(10)) // app90 to app99

		})

		It("should limit the result", func() {
			listOpts := store.NewListOpts()
			listOpts.Limit = 10

			list, err := objStore.List(ctx, listOpts)
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(10))
			// foo, foo0, foo1, foo10, foo100, foo101, foo102, foo103, foo104, foo105
			Expect(list.Links.Next).To(Equal("default/foo106"))
		})

		It("should support cursor", func() {
			listOpts := store.NewListOpts()
			listOpts.Limit = 10
			listOpts.Cursor = "default/foo107"

			list, err := objStore.List(ctx, listOpts)
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(10))
			// foo107, foo108, foo109, foo11, foo110, foo111, foo112, foo113, foo114, foo115
			Expect(list.Links.Next).To(Equal("default/foo116"))
		})

	})

	Context("EventHandler", func() {
		fakeClient := fake.NewSimpleDynamicClient(scheme, NewUnstructured("foo"))
		objStore := NewOrDie[*unstructured.Unstructured](ctx, StoreOpts{
			Client:       fakeClient,
			GVR:          gvr,
			GVK:          gvk,
			AllowedSorts: nil,
		})

		BeforeEach(func() {
			fakeClient.ClearActions()
		})

		It("should handle create event", func() {
			Expect(objStore.(*InmemoryObjectStore[*unstructured.Unstructured]).OnCreate(ctx, NewUnstructured("foo"))).To(Succeed())

			obj, err := objStore.Get(ctx, "default", "foo")
			Expect(err).ToNot(HaveOccurred())
			Expect(obj.GetName()).To(Equal("foo"))
		})

		It("should handle delete event", func() {
			Expect(objStore.(*InmemoryObjectStore[*unstructured.Unstructured]).OnDelete(ctx, NewUnstructured("foo"))).To(Succeed())

			obj, err := objStore.Get(ctx, "default", "foo")
			Expect(err).To(HaveOccurred())
			Expect(problems.IsNotFound(err)).To(BeTrue())
			Expect(obj).To(BeNil())
		})
	})

	Context("Kubernetes error mapping", func() {

		It("should handle nil errors", func() {
			err := mapErrorToProblem(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should map any error", func() {
			err := fmt.Errorf("foo")
			problem := mapErrorToProblem(err)
			Expect(problem.Code()).To(Equal(http.StatusInternalServerError))
		})

		It("should map not found error", func() {
			err := apierrors.NewNotFound(schema.GroupResource{}, "foo")
			problem := mapErrorToProblem(err)
			Expect(problems.IsNotFound(problem)).To(BeTrue())
		})

		It("should map conflict error", func() {
			err := apierrors.NewConflict(schema.GroupResource{}, "foo", fmt.Errorf("bar"))
			problem := mapErrorToProblem(err)
			Expect(problem.Code()).To(Equal(http.StatusConflict))
		})

		It("should map bad-request error", func() {
			err := apierrors.NewBadRequest("foo")
			err.ErrStatus.Details = &metav1.StatusDetails{
				Causes: []metav1.StatusCause{
					{
						Field:   "metadata.name",
						Message: "name is required",
					},
				},
			}
			problem := mapErrorToProblem(err)
			Expect(problem.Code()).To(Equal(http.StatusBadRequest))
		})

		It("should map unknown errors", func() {
			err := apierrors.NewForbidden(schema.GroupResource{}, "foo", fmt.Errorf("bar"))
			problem := mapErrorToProblem(err)
			Expect(problem.Code()).To(Equal(http.StatusInternalServerError))
		})
	})
})

package inmemory

import (
	"context"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/telekom/controlplane-mono/common-server/internal/informer"
	"github.com/telekom/controlplane-mono/common-server/pkg/problems"
	"github.com/telekom/controlplane-mono/common-server/pkg/store"
	"github.com/telekom/controlplane-mono/common-server/pkg/store/inmemory/filter"
	"github.com/telekom/controlplane-mono/common-server/pkg/store/inmemory/patch"
	"github.com/telekom/controlplane-mono/common-server/pkg/utils"
	"github.com/tidwall/gjson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var _ store.ObjectStore[store.Object] = &InmemoryObjectStore[store.Object]{}
var _ informer.EventHandler = &InmemoryObjectStore[store.Object]{}

type StoreOpts struct {
	Client       dynamic.Interface
	GVR          schema.GroupVersionResource
	GVK          schema.GroupVersionKind
	AllowedSorts []string
}

type InmemoryObjectStore[T store.Object] struct {
	ctx            context.Context
	log            logr.Logger
	gvr            schema.GroupVersionResource
	gvk            schema.GroupVersionKind
	k8sClient      dynamic.NamespaceableResourceInterface
	informer       *informer.Informer
	db             *badger.DB
	allowedSorts   []string
	sortValueCache sync.Map
}

func newDbOrDie(log logr.Logger) *badger.DB {
	opts := badger.DefaultOptions("").WithInMemory(true)
	opts.IndexCacheSize = 100 << 20
	opts.Logger = NewLoggerShim(log)
	db, err := badger.Open(opts)
	if err != nil {
		panic(errors.Wrap(err, "failed to create in-memory store"))
	}
	return db
}

func NewOrDie[T store.Object](ctx context.Context, storeOpts StoreOpts) store.ObjectStore[T] {
	store := &InmemoryObjectStore[T]{
		ctx:       ctx,
		log:       logr.FromContextOrDiscard(ctx),
		gvr:       storeOpts.GVR,
		gvk:       storeOpts.GVK,
		k8sClient: storeOpts.Client.Resource(storeOpts.GVR),
	}
	var err error
	store.db = newDbOrDie(store.log)
	store.informer = informer.New(ctx, store.gvr, storeOpts.Client, store)

	if err = store.informer.Start(); err != nil {
		panic(errors.Wrap(err, "failed to start informer"))
	}

	return store
}

func (s *InmemoryObjectStore[T]) Info() (schema.GroupVersionResource, schema.GroupVersionKind) {
	return s.gvr, s.gvk
}

func (s *InmemoryObjectStore[T]) Ready() bool {
	return s.informer.Ready()
}

func (s *InmemoryObjectStore[T]) Get(ctx context.Context, namespace, name string) (result T, err error) {
	key := newKey(namespace, name)
	err = s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return problems.NotFound(key)
			}
			return errors.Wrapf(err, "failed to get item %s", key)
		}
		if item == nil {
			return problems.NotFound(key)
		}

		return item.Value(func(val []byte) error {
			return utils.Unmarshal(val, &result)
		})
	})

	return result, err
}

func (s *InmemoryObjectStore[T]) List(ctx context.Context, listOpts store.ListOpts) (result *store.ListResponse[T], err error) {
	s.log.V(1).Info("list", "limit", listOpts.Limit, "cursor", listOpts.Cursor)

	hasFilters := len(listOpts.Filters) > 0

	result = &store.ListResponse[T]{
		Items: make([]T, listOpts.Limit),
	}

	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(listOpts.Prefix)
	if !hasFilters {
		opts.PrefetchValues = true
	} else {
		opts.PrefetchValues = false
	}

	var filterFunc filter.FilterFunc

	if hasFilters {
		filterFunc = filter.NewFilterFuncs(listOpts.Filters)
	} else {
		filterFunc = filter.NopFilter
	}

	startCursor := []byte(listOpts.Cursor)
	prefix := opts.Prefix
	startKey := prefix
	if len(startCursor) > 0 {
		startKey = startCursor
	}

	limit := listOpts.Limit
	iterNum := 0

	err = s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(startKey); it.ValidForPrefix(prefix); it.Next() {
			var err error
			var value []byte
			key := string(it.Item().Key())

			if iterNum >= limit {
				s.log.V(1).Info("limit reached", "limit", limit, "cursor", key)
				result.Links.Next = key
				return nil
			}

			if !hasFilters {
				value, err = it.Item().ValueCopy(nil)
				if err != nil {
					return err
				}

			} else {
				err = it.Item().Value(func(val []byte) error {
					if !filterFunc(val) {
						return nil
					}
					value = val
					return nil
				})
				if err != nil {
					return err
				}
			}

			if value != nil {
				err = utils.Unmarshal(value, &result.Items[iterNum])
				if err != nil {
					return errors.Wrap(err, "invalid object")
				}
				if result.Links.Self == "" {
					result.Links.Self = key
				}
				iterNum++
			}
		}
		return nil
	})

	if iterNum < listOpts.Limit {
		result.Links.Next = ""
		// truncate the list
		result.Items = result.Items[:iterNum]
	}

	return result, err
}

func (s *InmemoryObjectStore[T]) Delete(ctx context.Context, namespace, name string) error {
	err := s.k8sClient.Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return mapErrorToProblem(err)
	}
	return nil
}

func (s *InmemoryObjectStore[T]) CreateOrReplace(ctx context.Context, in T) error {
	if in.GetName() == "" {
		return problems.ValidationError("metadata.name", "name is required")
	}
	if in.GetNamespace() == "" {
		return problems.ValidationError("metadata.namespace", "namespace is required")
	}

	obj, err := convertToUnstructured(in)
	if err != nil {
		return errors.Wrap(err, "failed to convert object")
	}

	oldObj, err := s.Get(ctx, obj.GetNamespace(), obj.GetName())
	if err != nil && !problems.IsNotFound(err) {
		return err
	}

	if problems.IsNotFound(err) {
		s.log.Info("creating object", "namespace", obj.GetNamespace(), "name", obj.GetName())
		obj.GetObjectKind().SetGroupVersionKind(s.gvk)

		// check if not found
		obj, err = s.k8sClient.Namespace(obj.GetNamespace()).Create(ctx, obj, metav1.CreateOptions{
			FieldValidation: "Strict",
		})
		if err != nil {
			return errors.Wrap(mapErrorToProblem(err), "failed to create object")
		}
		return s.OnCreate(ctx, obj)
	}

	obj.SetResourceVersion(oldObj.GetResourceVersion())
	obj, err = s.k8sClient.Namespace(obj.GetNamespace()).Update(ctx, obj, metav1.UpdateOptions{
		FieldValidation: "Strict",
	})
	if err != nil {
		return errors.Wrap(mapErrorToProblem(err), "failed to update object")
	}

	return s.OnUpdate(ctx, obj)
}

func (s *InmemoryObjectStore[T]) Patch(ctx context.Context, namespace, name string, ops ...store.Patch) (obj T, err error) {

	if len(ops) == 0 {
		return obj, errors.New("no patch operations provided")
	}

	var value []byte
	patchFunc := patch.NewPatchFuncs(ops)

	err = s.db.View(func(txn *badger.Txn) error {
		key := newKey(namespace, name)
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}
		if item == nil {
			return nil // Not found
		}

		err = item.Value(func(val []byte) error {
			value = val
			return nil
		})

		if err != nil {
			return errors.Wrap(err, "failed to get value")
		}

		return nil
	})

	if err != nil {
		return obj, errors.Wrap(err, "failed to get value")
	}

	if value == nil {
		return obj, errors.New("object not found")
	}

	value, err = patchFunc(value)
	if err != nil {
		return obj, errors.Wrap(err, "failed to patch object")
	}

	err = utils.Unmarshal(value, &obj)
	if err != nil {
		return obj, errors.Wrap(err, "failed to unmarshal patched object")
	}
	return obj, s.CreateOrReplace(ctx, obj)
}

func (s *InmemoryObjectStore[T]) OnCreate(ctx context.Context, obj *unstructured.Unstructured) error {
	return s.OnUpdate(ctx, obj)
}

func (s *InmemoryObjectStore[T]) OnUpdate(ctx context.Context, obj *unstructured.Unstructured) error {
	key := calculateKey(obj)
	informer.SanitizeObject(obj)

	data, err := utils.Marshal(obj.Object)
	if err != nil {
		return errors.Wrap(err, "invalid object")
	}
	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
	if err != nil {
		return err
	}

	s.sortValueCache.Range(func(k, v any) bool {
		sp := k.(string)
		m := v.(*sync.Map)
		value := gjson.GetBytes(data, sp)
		m.Store(key, value.Value())
		s.log.V(1).Info("cached sort value", "key", key, "sortPath", sp, "value", value.Value())
		return true
	})
	return nil
}

func (s *InmemoryObjectStore[T]) OnDelete(ctx context.Context, obj *unstructured.Unstructured) error {
	key := calculateKey(obj)

	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	if err != nil {
		return err
	}
	s.sortValueCache.Range(func(k, v any) bool {
		m := v.(*sync.Map)
		m.Delete(key)
		s.log.V(1).Info("deleted cached sort value", "key", key, "sortPath", k)
		return true
	})
	return nil
}

func mapErrorToProblem(err error) problems.Problem {
	if err == nil {
		return nil
	}
	apiStatus, ok := err.(apierrors.APIStatus)
	if !ok {
		return problems.NewProblemOfError(err)
	}
	status := apiStatus.Status()

	switch status.Code {
	case 404:
		return problems.NotFound(status.Kind)
	case 400, 422:
		if status.Details == nil {
			return problems.BadRequest(status.Message)
		}
		causes := status.Details.Causes
		if len(causes) == 0 {
			return problems.BadRequest(status.Message)
		}

		fields := make(map[string]string, len(causes))
		for _, cause := range causes {
			if _, ok := fields[cause.Field]; ok {
				fields[cause.Field] += ", " + cause.Message
				continue
			}
			fields[cause.Field] = cause.Message
		}
		return problems.ValidationErrors(fields)

	case 409:
		return problems.Conflict(status.Message)

	default:
		return problems.NewProblemOfError(err)
	}
}

func calculateKey(obj store.Object) string {
	return newKey(obj.GetNamespace(), obj.GetName())
}

func newKey(namespace, name string) string {
	return namespace + "/" + name
}

func convertToUnstructured(obj any) (*unstructured.Unstructured, error) {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: u}, nil
}

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

	"github.com/telekom/controlplane-mono/common/pkg/config"
	commonController "github.com/telekom/controlplane-mono/common/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	identityv1 "github.com/telekom/controlplane-mono/identity/api/v1"
	clientHandler "github.com/telekom/controlplane-mono/identity/internal/handler/client"
)

// ClientReconciler reconciles a Client object
type ClientReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	commonController.Controller[*identityv1.Client]
}

// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=identity.cp.ei.telekom.de,resources=clients,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=identity.cp.ei.telekom.de,resources=clients/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=identity.cp.ei.telekom.de,resources=clients/finalizers,verbs=update
// +kubebuilder:rbac:groups=identity.cp.ei.telekom.de,resources=realms,verbs=get;list;watch;create;update;patch;delete

func (r *ClientReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Controller.Reconcile(ctx, req, &identityv1.Client{})
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClientReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("client-controller")
	r.Controller = commonController.NewController(&clientHandler.HandlerClient{}, r.Client, r.Recorder)

	return ctrl.NewControllerManagedBy(mgr).
		For(&identityv1.Client{}).
		Watches(&identityv1.Realm{},
			handler.EnqueueRequestsFromMapFunc(r.mapRealmObjToIdentityClient),
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 10,
			RateLimiter:             workqueue.DefaultTypedItemBasedRateLimiter[reconcile.Request](),
		}).
		Complete(r)
}

// mapRealmObjToIdentityClient maps identity realm object to reconcile requests.
func (r *ClientReconciler) mapRealmObjToIdentityClient(ctx context.Context, obj client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	realm, ok := obj.(*identityv1.Realm)
	if !ok {
		logger.V(0).Info("object is not a Realm")
		return nil
	}

	list := &identityv1.ClientList{}
	err := r.Client.List(ctx, list, client.MatchingLabels{
		config.EnvironmentLabelKey: realm.Labels[config.EnvironmentLabelKey],
	})
	if err != nil {
		logger.Error(err, "failed to list clients")
		return nil
	}

	requests := make([]reconcile.Request, 0, len(list.Items))
	for _, item := range list.Items {
		if realm.UID == item.UID {
			continue
		}
		requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&item)})
	}

	return requests
}

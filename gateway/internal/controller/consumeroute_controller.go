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
	"github.com/telekom/controlplane-mono/common/pkg/controller"
	"github.com/telekom/controlplane-mono/common/pkg/types"
	gatewayv1 "github.com/telekom/controlplane-mono/gateway/api/v1"
	consumeroute_handler "github.com/telekom/controlplane-mono/gateway/internal/handler/consumeroute"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ConsumeRouteReconciler reconciles a ConsumeRoute object
type ConsumeRouteReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	controller.Controller[*gatewayv1.ConsumeRoute]
}

// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=gateway.cp.ei.telekom.de,resources=consumeroutes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.cp.ei.telekom.de,resources=consumeroutes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gateway.cp.ei.telekom.de,resources=consumeroutes/finalizers,verbs=update

func (r *ConsumeRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Controller.Reconcile(ctx, req, &gatewayv1.ConsumeRoute{})
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConsumeRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("consumeroute-controller")
	r.Controller = controller.NewController(&consumeroute_handler.ConsumeRouteHandler{}, r.Client, r.Recorder)

	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1.ConsumeRoute{}).
		Watches(&gatewayv1.Route{},
			handler.EnqueueRequestsFromMapFunc(r.mapRouteToConsumeRoute),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}

func (r *ConsumeRouteReconciler) mapRouteToConsumeRoute(ctx context.Context, obj client.Object) []reconcile.Request {
	// ensure its actually a Realm
	route, ok := obj.(*gatewayv1.Route)
	if !ok {
		return nil
	}
	if route.Labels == nil {
		return nil
	}

	listOpts := []client.ListOption{
		client.MatchingFields{
			IndexFieldSpecRoute: types.ObjectRefFromObject(route).String(),
		},
		client.MatchingLabels{
			config.EnvironmentLabelKey: route.Labels[config.EnvironmentLabelKey],
		},
	}

	list := gatewayv1.ConsumeRouteList{}
	if err := r.List(ctx, &list, listOpts...); err != nil {
		return nil
	}

	requests := make([]reconcile.Request, len(list.Items))
	for i, item := range list.Items {
		if item.Spec.Route.Equals(route) {
			requests[i] = reconcile.Request{NamespacedName: client.ObjectKey{Name: item.Name, Namespace: item.Namespace}}
		}
	}

	return requests
}

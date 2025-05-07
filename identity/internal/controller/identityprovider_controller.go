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

	commonController "github.com/telekom/controlplane-mono/common/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	identityv1 "github.com/telekom/controlplane-mono/identity/api/v1"
	identityproviderHandler "github.com/telekom/controlplane-mono/identity/internal/handler/identityprovider"
)

// IdentityProviderReconciler reconciles a IdentityProvider object
type IdentityProviderReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	commonController.Controller[*identityv1.IdentityProvider]
}

// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=identity.cp.ei.telekom.de,resources=identityproviders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=identity.cp.ei.telekom.de,resources=identityproviders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=identity.cp.ei.telekom.de,resources=identityproviders/finalizers,verbs=update

func (r *IdentityProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Controller.Reconcile(ctx, req, &identityv1.IdentityProvider{})
}

// SetupWithManager sets up the controller with the Manager.
func (r *IdentityProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("identityprovider-controller")
	r.Controller = commonController.NewController(&identityproviderHandler.HandlerIdentityProvider{}, r.Client, r.Recorder)

	// TODO CreateOrUpdate realms in keycloak

	return ctrl.NewControllerManagedBy(mgr).
		For(&identityv1.IdentityProvider{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 10,
			RateLimiter:             workqueue.DefaultTypedItemBasedRateLimiter[reconcile.Request](),
		}).
		Complete(r)
}

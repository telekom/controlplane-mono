package client

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/telekom/controlplane-mono/common/pkg/handler"
	"github.com/telekom/controlplane-mono/common/pkg/util/contextutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"

	identityv1 "github.com/telekom/controlplane-mono/identity/api/v1"
	realmHandler "github.com/telekom/controlplane-mono/identity/internal/handler/realm"
	"github.com/telekom/controlplane-mono/identity/pkg/keycloak"
	secrets "github.com/telekom/controlplane-mono/secret-manager/pkg/api"
)

var _ handler.Handler[*identityv1.Client] = &HandlerClient{}

type HandlerClient struct{}

func (h *HandlerClient) CreateOrUpdate(ctx context.Context, client *identityv1.Client) (err error) {
	logger := log.FromContext(ctx)
	if client == nil {
		return fmt.Errorf("client is nil")
	}

	SetStatusProcessing(&client.Status, client)

	// Get secret-values from secret-manager
	client.Spec.ClientSecret, err = secrets.Get(ctx, client.Spec.ClientSecret)
	if err != nil {
		return errors.Wrap(err, "failed to get client secret from secret-manager")
	}

	realm, err := realmHandler.GetRealmByName(ctx, client.Spec.Realm)
	if err != nil {
		if apierrors.IsNotFound(err) {
			contextutil.RecorderFromContextOrDie(ctx).
				Eventf(client, "Warning", "RealmNotFound",
					"Realm '%s' not found", client.Spec.Realm.String())
			SetStatusBlocked(&client.Status, client)
			return nil
		}
		return err
	}
	realmStatus := realmHandler.ObfuscateRealm(realm.Status)
	logger.V(0).Info("Found Realm", "realm", realmStatus)

	var clientStatus = MapToClientStatus(&realm.Status)
	err = realmHandler.ValidateRealmStatus(&realm.Status)
	if err != nil {
		contextutil.RecorderFromContextOrDie(ctx).
			Eventf(client, "Warning", "RealmNotValid",
				"Realm '%s' not valid", client.Spec.Realm.String())
		SetStatusWaiting(&client.Status, client)
		return errors.Wrap(err, "❌ failed to validate realm")
	}

	realmStatus.AdminPassword, err = secrets.Get(ctx, realmStatus.AdminPassword)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve password from secret manager")
	}
	realmClient, err := keycloak.GetClientFor(realm.Status)
	if err != nil {
		return errors.Wrap(err, "❌ failed to get keycloak client")
	}

	err = realmClient.CreateOrUpdateRealmClient(ctx, realm, client)
	if err != nil {
		return errors.Wrap(err, "❌ failed to create or update client")
	}

	SetStatusReady(&clientStatus, client)
	var message = fmt.Sprintf("✅ RealmClient %s is ready", client.Spec.ClientId)
	logger.V(1).Info(message, "IssuerUrl", clientStatus.IssuerUrl)

	return nil
}

func (h *HandlerClient) Delete(ctx context.Context, obj *identityv1.Client) error {
	return nil
}

package realm

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/telekom/controlplane-mono/common/pkg/client"
	"github.com/telekom/controlplane-mono/common/pkg/condition"
	common "github.com/telekom/controlplane-mono/common/pkg/types"
	"k8s.io/apimachinery/pkg/api/meta"

	identityv1 "github.com/telekom/controlplane-mono/identity/api/v1"
)

func GetRealmByName(ctx context.Context, realmRef *common.ObjectRef) (*identityv1.Realm, error) {
	clientFromContext := client.ClientFromContextOrDie(ctx)

	realm := &identityv1.Realm{}
	err := clientFromContext.Get(context.Background(), realmRef.K8s(), realm)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get realm %s", realmRef.String())
	}
	if !meta.IsStatusConditionTrue(realm.GetConditions(), condition.ConditionTypeReady) {
		return nil, nil
	}
	return realm, nil
}

func ValidateRealmStatus(realmStatus *identityv1.RealmStatus) error {
	if realmStatus == nil {
		return fmt.Errorf("realmStatus is nil")
	}
	if realmStatus.IssuerUrl == "" {
		return fmt.Errorf("realmStatus.IssuerUrl is empty")
	}
	if realmStatus.AdminClientId == "" {
		return fmt.Errorf("realmStatus.AdminClientId is empty")
	}
	if realmStatus.AdminUserName == "" {
		return fmt.Errorf("realmStatus.AdminUserName is empty")
	}
	if realmStatus.AdminPassword == "" {
		return fmt.Errorf("realmStatus.AdminPassword is empty")
	}
	if realmStatus.AdminUrl == "" {
		return fmt.Errorf("realmStatus.AdminUrl is empty")
	}
	if realmStatus.AdminTokenUrl == "" {
		return fmt.Errorf("realmStatus.AdminTokenUrl is empty")
	}
	return nil
}

func ObfuscateRealm(status identityv1.RealmStatus) identityv1.RealmStatus {
	// Create a copy of the status to avoid modifying the original
	obfuscatedStatus := status

	// Obfuscate sensitive fields
	if obfuscatedStatus.AdminUserName != "" {
		obfuscatedStatus.AdminUserName = "****"
	}
	if obfuscatedStatus.AdminPassword != "" {
		obfuscatedStatus.AdminPassword = "****"
	}

	return obfuscatedStatus
}

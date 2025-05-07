package mapper

import (
	"k8s.io/utils/ptr"

	identityv1 "github.com/telekom/controlplane-mono/identity/api/v1"
	"github.com/telekom/controlplane-mono/identity/pkg/api"
)

func MapToRealmRepresentation(realm *identityv1.Realm) api.RealmRepresentation {
	return api.RealmRepresentation{
		Enabled: ptr.To(true),
		Realm:   &realm.Name,
	}
}

func CompareRealmRepresentation(existingRealm, newRealm *api.RealmRepresentation) bool {
	return *existingRealm.Realm == *newRealm.Realm &&
		*existingRealm.Enabled == *newRealm.Enabled
}

func MergeRealmRepresentation(existingRealm, newRealm *api.RealmRepresentation) *api.RealmRepresentation {
	existingRealm.Enabled = newRealm.Enabled
	existingRealm.Realm = newRealm.Realm
	return existingRealm
}

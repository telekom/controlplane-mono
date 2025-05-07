package realm

import (
	"github.com/telekom/controlplane-mono/common/pkg/config"
	"github.com/telekom/controlplane-mono/common/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	identityv1 "github.com/telekom/controlplane-mono/identity/api/v1"
)

func NewRealmSpec(identityProviderName string, namespace string) *identityv1.RealmSpec {
	return &identityv1.RealmSpec{
		IdentityProvider: &types.ObjectRef{
			Name:      identityProviderName,
			Namespace: namespace,
		},
	}
}

func NewRealmMeta(name string, namespace string, environment string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels: map[string]string{
			config.EnvironmentLabelKey: environment,
		},
	}
}

func NewRealm(name string, namespace string, environment string, identityProviderName string) *identityv1.Realm {
	return &identityv1.Realm{
		ObjectMeta: *NewRealmMeta(name, namespace, environment),
		Spec:       *NewRealmSpec(identityProviderName, namespace),
	}
}

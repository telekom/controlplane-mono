package client

import (
	"github.com/telekom/controlplane-mono/common/pkg/config"
	"github.com/telekom/controlplane-mono/common/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	identityv1 "github.com/telekom/controlplane-mono/identity/api/v1"
)

func NewClientSpec(realmName string, namespace string) *identityv1.ClientSpec {
	return &identityv1.ClientSpec{
		Realm: &types.ObjectRef{
			Name:      realmName,
			Namespace: namespace,
		},
		ClientId:     "test-client",
		ClientSecret: "test-secret",
	}
}

func NewClientMeta(name string, namespace string, environment string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels: map[string]string{
			config.EnvironmentLabelKey: environment,
		},
	}
}

func NewClient(name string, namespace string, environment string, realmName string) *identityv1.Client {
	return &identityv1.Client{
		ObjectMeta: *NewClientMeta(name, namespace, environment),
		Spec:       *NewClientSpec(realmName, namespace),
	}
}

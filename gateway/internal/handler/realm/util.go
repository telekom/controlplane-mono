package realm

import (
	"context"

	"github.com/pkg/errors"
	"github.com/telekom/controlplane-mono/common/pkg/client"
	"github.com/telekom/controlplane-mono/common/pkg/condition"
	"github.com/telekom/controlplane-mono/common/pkg/types"
	v1 "github.com/telekom/controlplane-mono/gateway/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
)

func GetRealmByRef(ctx context.Context, ref types.ObjectRef) (bool, *v1.Realm, error) {
	client := client.ClientFromContextOrDie(ctx)

	realm := &v1.Realm{}
	err := client.Get(ctx, ref.K8s(), realm)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, errors.Wrapf(err, "failed to get realm %s", ref.String())

	}

	if !meta.IsStatusConditionTrue(realm.GetConditions(), condition.ConditionTypeReady) {
		return false, realm, nil
	}
	return true, realm, nil
}

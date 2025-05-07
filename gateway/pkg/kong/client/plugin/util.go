package plugin

import (
	"github.com/telekom/controlplane-mono/gateway/pkg/kong/client"
)

func As[T client.CustomPlugin](s client.CustomPlugin, t *T) bool {
	if st, ok := s.(T); ok {
		*t = st
		return true
	}
	return false
}

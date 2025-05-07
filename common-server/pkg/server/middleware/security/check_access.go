package security

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"github.com/gofiber/fiber/v2"
	"github.com/telekom/controlplane-mono/common-server/pkg/problems"
	"github.com/telekom/controlplane-mono/common-server/pkg/server/middleware/util"
)

var accessDenied = problems.Forbidden("Access Denied", "Access to requested resource not allowed")

// ExpectedNamespaceFunc must return the expected namespace for the provided BusinessContext
type NamespaceFunc func(*BusinessContext) string

var defaultExpectedNamespace NamespaceFunc = func(bCtx *BusinessContext) string {
	switch bCtx.ClientType {
	case ClientTypeAdmin:
		return bCtx.Environment + "--"
	case ClientTypeGroup:
		return bCtx.Environment + "--" + bCtx.Group + "--"
	default:
		return bCtx.Environment + "--" + bCtx.Group + "--" + bCtx.Team
	}
}

var defaultPrefix = func(nsf NamespaceFunc) NamespaceFunc {
	return func(bCtx *BusinessContext) string {
		ns := nsf(bCtx)

		if bCtx.ClientType == ClientTypeTeam {
			return ns + "/"
		}
		return ns
	}
}

type CheckAccessOpts struct {
	PathParamKey      string
	ExpectedNamespace NamespaceFunc
	// Prefix is used to calculate the prefix that is used in the context
	// It is expected that this prefix is then used to determine the access to the store
	// The store uses a key-format `<namespace>/<name>` to store resources.
	// At most the prefix should match the namespace part of the key
	Prefix NamespaceFunc
}

func WithExpectedNamespaceFunc(f NamespaceFunc) Option[*CheckAccessOpts] {
	return func(o *CheckAccessOpts) {
		o.ExpectedNamespace = f
	}
}

func WithPrefixFunc(f NamespaceFunc) Option[*CheckAccessOpts] {
	return func(o *CheckAccessOpts) {
		o.Prefix = f
	}
}

func WithPathParamKey(key string) Option[*CheckAccessOpts] {
	return func(o *CheckAccessOpts) {
		o.PathParamKey = key
	}
}

/*
NewCheckAccessMiddlewareWithOpts creates a new middleware that checks if the client has access to the requested resource

# It is expected that `business_context` middleware is executed before this middleware

The middleware checks the client's context and access rights to determine if the client has access to the requested resource.

As this middleware depends on path-params it has to be configured on the route level
```go
app.Get("/api/v1/foos/:namespace/:name", NewCheckAccessMiddlewareWithOpts(), handler)
```
*/
func NewCheckAccessMiddlewareWithOpts(opts ...Option[*CheckAccessOpts]) fiber.Handler {
	mwOpts := CheckAccessOpts{
		PathParamKey:      "namespace",
		ExpectedNamespace: defaultExpectedNamespace,
	}

	for _, f := range opts {
		f(&mwOpts)
	}
	if mwOpts.Prefix == nil {
		mwOpts.Prefix = defaultPrefix(mwOpts.ExpectedNamespace)
	}

	return func(c *fiber.Ctx) error {
		namespace := c.Params(mwOpts.PathParamKey)
		bCtx, ok := util.NotNilOfType[*BusinessContext](c.Locals("businessContext"))
		if !ok {
			return c.Status(accessDenied.Code()).JSON(accessDenied, "application/problem+json")
		}
		expectedNamespace := mwOpts.ExpectedNamespace(bCtx)

		allow := false
		if namespace == "" {
			allow = CheckGlobalRequest(c, bCtx)
			writeLog(c, bCtx, "global", allow)
		} else {
			allow = CheckNamespacedRequest(c, namespace, expectedNamespace, bCtx)
			writeLog(c, bCtx, "namespaced", allow)
		}

		if allow {
			prefix := mwOpts.Prefix(bCtx)
			c.Locals("prefix", prefix)

			ctx := c.UserContext()
			ctx = context.WithValue(ctx, prefixKey, prefix)
			ctx = ToContext(ctx, bCtx)
			c.SetUserContext(ctx)
			return c.Next()
		}
		return c.Status(accessDenied.Code()).JSON(accessDenied, "application/problem+json")
	}
}

func IsReadyOnlyRequest(c *fiber.Ctx) bool {
	return c.Method() == "GET" || c.Method() == "HEAD"
}

// CheckNamespacedRequest performs access checks for namespaced requests
// Namespaced requests are requests that have a direct resource reference
// e.g. requests for a specific resource `GET /api/v1/foos/<namespace>/<name>`
// It does to by matching the prefix of the namespace with the client's context
// and checking if the client has the required access rights
func CheckNamespacedRequest(c *fiber.Ctx, actual, expected string, bCtx *BusinessContext) (allow bool) {
	if bCtx.ClientType == ClientTypeAdmin {
		if IsReadyOnlyRequest(c) {
			allow = true
		} else {
			allow = bCtx.AccessType.IsWrite()
		}
	}

	groupAllowed := strings.HasPrefix(actual, expected)
	if bCtx.ClientType == ClientTypeGroup {
		if IsReadyOnlyRequest(c) {
			allow = groupAllowed && bCtx.AccessType.IsRead()
		} else {
			allow = groupAllowed && bCtx.AccessType.IsWrite()
		}
	}

	teamAllowed := actual == expected
	if bCtx.ClientType == ClientTypeTeam {
		if IsReadyOnlyRequest(c) {
			allow = teamAllowed && bCtx.AccessType.IsRead()
		} else {
			allow = teamAllowed && bCtx.AccessType.IsWrite()
		}
	}
	return allow

}

// CheckGlobalRequest performs access checks for global requests
// Global requests are requests are all requests that do not have a direct resource reference
// e.g. requests for to list resource `GET /api/v1/foos`
func CheckGlobalRequest(c *fiber.Ctx, bCtx *BusinessContext) (allow bool) {
	if bCtx.ClientType == ClientTypeAdmin {
		if IsReadyOnlyRequest(c) {
			allow = true
		} else {
			allow = bCtx.AccessType == AccessTypeReadWrite
		}
	}

	if bCtx.ClientType == ClientTypeGroup {
		if IsReadyOnlyRequest(c) {
			allow = bCtx.AccessType.IsRead()
		} else {
			allow = bCtx.AccessType.IsWrite()
		}
	}

	if bCtx.ClientType == ClientTypeTeam {
		if IsReadyOnlyRequest(c) {
			allow = bCtx.AccessType.IsRead()
		} else {
			allow = bCtx.AccessType.IsWrite()
		}
	}
	return allow
}

func writeLog(c *fiber.Ctx, bCtx *BusinessContext, requestType string, allow bool) {
	logArgs := []interface{}{"method", c.Method(), "type", requestType, "env", bCtx.Environment, "group", bCtx.Group, "team", bCtx.Team, "clientType", bCtx.ClientType, "accessType", bCtx.AccessType}
	if allow {
		logr.FromContextOrDiscard(c.UserContext()).Info("Access granted", logArgs...)
	} else {
		logr.FromContextOrDiscard(c.UserContext()).Info("Access denied", logArgs...)
	}
}

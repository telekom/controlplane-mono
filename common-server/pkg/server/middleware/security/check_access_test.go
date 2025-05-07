package security

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/telekom/controlplane-mono/common-server/pkg/server/middleware/security/mock"
	"github.com/telekom/controlplane-mono/common-server/pkg/server/middleware/util"
)

var handlerMock = func(c *fiber.Ctx) error {
	bCtx, ok := util.NotNilOfType[*BusinessContext](c.Locals("businessContext"))
	if !ok {
		return c.SendString("Invalid authorization context")
	}
	prefix, ok := c.Locals("prefix").(string)
	if !ok {
		prefix = ""
	}
	return c.SendString("prefix:" + prefix + ", group:" + bCtx.Group + ", team:" + bCtx.Team)
}

type TestCase struct {
	Method         string
	AccessToken    string
	Namespace      string
	ExpectedStatus int
	ExpectedBody   string
	Description    string // Added description field
}

var env = "test"

func TestCheckAccess(t *testing.T) {
	var TestCases = []TestCase{
		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", nil),
			Namespace:      env + "--group--team",
			ExpectedStatus: 200,
			ExpectedBody:   "prefix:test--group--team/, group:group, team:team",
			Description:    "Valid token with group and team, matching namespace",
		},
		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("test", "group", "", nil),
			Namespace:      env + "--group",
			ExpectedStatus: 403,
			ExpectedBody:   "Invalid authorization context",
			Description:    "Valid token with group but no team, namespace mismatch",
		},
		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("test", "", "", nil),
			Namespace:      env + "--group",
			ExpectedStatus: 403,
			ExpectedBody:   "Invalid authorization context",
			Description:    "Valid token with no group and no team, namespace mismatch",
		},
		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("", "", "", nil),
			Namespace:      env + "--group",
			ExpectedStatus: 403,
			ExpectedBody:   "Missing field 'env'",
			Description:    "Token missing 'env' field",
		},
		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", []string{"team:obfuscated"}),
			Namespace:      env + "--group--team",
			ExpectedStatus: 200,
			ExpectedBody:   "prefix:test--group--team/, group:group, team:team",
			Description:    "Valid token with obfuscated team scope, matching namespace",
		},
		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", []string{"team:read"}),
			Namespace:      env + "--othergroup",
			ExpectedStatus: 403,
			ExpectedBody:   "Access to requested resource not allowed",
			Description:    "Valid token with team scope, namespace mismatch",
		},
		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", []string{"group:read"}),
			Namespace:      env + "--group--foo",
			ExpectedStatus: 200,
			ExpectedBody:   "prefix:test--group--, group:group, team:team",
			Description:    "Valid token with group scope, matching namespace",
		},
		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", []string{"group:read"}),
			Namespace:      env + "--othergroup",
			ExpectedStatus: 403,
			ExpectedBody:   "Access to requested resource not allowed",
			Description:    "Valid token with group scope, namespace mismatch",
		},
		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", []string{"admin:read"}),
			Namespace:      env + "--othergroup",
			ExpectedStatus: 200,
			ExpectedBody:   "prefix:test--, group:group, team:team",
			Description:    "Valid token with admin scope, matching namespace",
		},
		{
			Method:         "POST",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", []string{"admin:all"}),
			Namespace:      env + "--othergroup",
			ExpectedStatus: 200,
			ExpectedBody:   "prefix:test--, group:group, team:team",
			Description:    "Valid token with admin all scope, matching namespace",
		},
		{
			Method:         "POST",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", []string{"admin:read"}),
			Namespace:      env + "--othergroup",
			ExpectedStatus: 403,
			ExpectedBody:   "Access to requested resource not allowed",
			Description:    "Valid token with admin read scope, namespace mismatch",
		},

		// Global

		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", nil),
			Namespace:      "",
			ExpectedStatus: 200,
			ExpectedBody:   "prefix:test--group--team/, group:group, team:team",
			Description:    "Valid token with group and team, no namespace",
		},
		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", []string{"group:read"}),
			Namespace:      "",
			ExpectedStatus: 200,
			ExpectedBody:   "prefix:test--group--, group:group, team:team",
			Description:    "Valid token with group scope, no namespace",
		},
		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", []string{"admin:read"}),
			Namespace:      "",
			ExpectedStatus: 200,
			ExpectedBody:   "prefix:test--, group:group, team:team",
			Description:    "Valid token with admin scope, no namespace",
		},
		{
			Method:         "POST",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", []string{"team:read"}),
			Namespace:      "",
			ExpectedStatus: 403,
			ExpectedBody:   "Access to requested resource not allowed",
			Description:    "Valid token with team scope, no namespace",
		},
		{
			Method:         "POST",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", []string{"team:all"}),
			Namespace:      "",
			ExpectedStatus: 200,
			ExpectedBody:   "prefix:test--group--team/, group:group, team:team",
			Description:    "Valid token with team all scope, no namespace",
		},
	}

	app := fiber.New()

	app.Use(mock.NewJWTMock())
	app.Use(NewBusinessCtxMiddlewareWithOpts(WithScopePrefix(""), WithDefaultScope("team:read")))
	checkAccess := NewCheckAccessMiddlewareWithOpts()

	app.All("/testauth/:namespace/:name", checkAccess, handlerMock)
	app.All("/testauth", checkAccess, handlerMock)

	for _, testCase := range TestCases {
		path := "/testauth"
		if testCase.Namespace != "" {
			path += "/" + testCase.Namespace + "/" + "foo"
		}
		req := httptest.NewRequest(testCase.Method, path, nil)
		req.Header.Set("Authorization", "Bearer "+testCase.AccessToken)

		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != testCase.ExpectedStatus {
			t.Fatalf("%s: Expected status %d, got %d", testCase.Description, testCase.ExpectedStatus, res.StatusCode)
		}

		var cmp func(a, b string) bool
		if testCase.ExpectedStatus != 200 {
			cmp = strings.Contains
		} else {
			cmp = strings.EqualFold
		}
		b, _ := io.ReadAll(res.Body)
		if !cmp(string(b), testCase.ExpectedBody) {
			t.Fatalf("%s: Expected body '%s', got '%s'", testCase.Description, testCase.ExpectedBody, string(b))
		}
	}

}

func TestCheckAccessOpts(t *testing.T) {
	var TestCases = []TestCase{
		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", nil),
			Namespace:      "group--team",
			ExpectedStatus: 403,
			ExpectedBody:   "Access to requested resource not allowed",
			Description:    "Valid token but namespace does not match expected format",
		},
		{
			Method:         "GET",
			AccessToken:    mock.NewMockAccessToken("test", "group", "team", nil),
			Namespace:      "foo--group--team",
			ExpectedStatus: 200,
			ExpectedBody:   "prefix:foo--group--team/, group:group, team:team",
			Description:    "Valid token with custom namespace matching expected format",
		},
	}

	app := fiber.New()

	app.Use(mock.NewJWTMock())
	app.Use(NewBusinessCtxMiddlewareWithOpts(WithScopePrefix(""), WithDefaultScope("team:read")))
	checkAccess := NewCheckAccessMiddlewareWithOpts(
		WithPathParamKey("customPathParam"),
		WithExpectedNamespaceFunc(func(bCtx *BusinessContext) string {
			return "foo--" + bCtx.Group + "--" + bCtx.Team
		}),
	)

	app.All("/testauth/:customPathParam", checkAccess, handlerMock)
	app.All("/testauth", checkAccess, handlerMock)

	for _, testCase := range TestCases {
		path := "/testauth/" + testCase.Namespace
		req := httptest.NewRequest(testCase.Method, path, nil)
		req.Header.Set("Authorization", "Bearer "+testCase.AccessToken)

		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != testCase.ExpectedStatus {
			t.Fatalf("%s: Expected status %d, got %d", testCase.Description, testCase.ExpectedStatus, res.StatusCode)
		}

		var cmp func(a, b string) bool
		if testCase.ExpectedStatus != 200 {
			cmp = strings.Contains
		} else {
			cmp = strings.EqualFold
		}

		b, _ := io.ReadAll(res.Body)
		if !cmp(string(b), testCase.ExpectedBody) {
			t.Fatalf("%s: Expected body '%s', got '%s'", testCase.Description, testCase.ExpectedBody, string(b))
		}
	}
}

func BenchmarkCheckAccess(b *testing.B) {

	app := fiber.New()

	app.Use(mock.NewJWTMock())
	app.Use(NewBusinessCtxMiddlewareWithOpts(WithScopePrefix(""), WithDefaultScope("team:read")))
	checkAccess := NewCheckAccessMiddlewareWithOpts()

	app.All("/testauth/:namespace/:name", checkAccess, handlerMock)
	app.All("/testauth", checkAccess, handlerMock)

	req := httptest.NewRequest("GET", "/testauth", nil)
	req.Header.Set("Authorization", "Bearer "+mock.NewMockAccessToken("test", "group", "team", nil))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res, err := app.Test(req)
		if err != nil {
			b.Fatal(err)
		}

		if res.StatusCode != 200 {
			b.Fatalf("Expected status 200, got %d", res.StatusCode)
		}
	}

}

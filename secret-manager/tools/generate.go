//go:build tools
// +build tools

package tools

import (
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
	_ "github.com/vektra/mockery/v2"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=server.yaml ../api/openapi.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=client.yaml ../api/openapi.yaml

//go:generate go run github.com/vektra/mockery/v2 --config=mockery.yaml

//go:generate go run github.com/vektra/mockery/v2 --config=mockery.api.yaml

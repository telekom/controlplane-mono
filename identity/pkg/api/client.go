package api

import (
	"context"
)

var _ KeycloakClient = &keycloakClient{}

type keycloakClient struct {
	KeycloakClient
}

type KeycloakClient interface {
	GetRealmWithResponse(ctx context.Context, realm string,
		reqEditors ...RequestEditorFn) (*GetRealmResponse, error)
	PutRealmWithResponse(ctx context.Context, realm string, body PutRealmJSONRequestBody,
		reqEditors ...RequestEditorFn) (*PutRealmResponse, error)
	PostWithResponse(ctx context.Context, body PostJSONRequestBody,
		reqEditors ...RequestEditorFn) (*PostResponse, error)

	GetRealmClientsWithResponse(ctx context.Context, realm string, params *GetRealmClientsParams,
		reqEditors ...RequestEditorFn) (*GetRealmClientsResponse, error)
	PutRealmClientsIdWithResponse(ctx context.Context, realm string, id string, body PutRealmClientsIdJSONRequestBody,
		reqEditors ...RequestEditorFn) (*PutRealmClientsIdResponse, error)
	PostRealmClientsWithResponse(ctx context.Context, realm string, body PostRealmClientsJSONRequestBody,
		reqEditors ...RequestEditorFn) (*PostRealmClientsResponse, error)
}

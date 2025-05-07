package middleware

import "github.com/golang-jwt/jwt/v5"

var _ jwt.Claims = (*ServiceAccountTokenClaims)(nil)

type Kubernetes struct {
	Namespace      string      `json:"namespace"`
	ServiceAccount NamedObject `json:"serviceaccount"`
	Pod            NamedObject `json:"pod"`
	Node           NamedObject `json:"node"`
}

type NamedObject struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}

type ServiceAccountTokenClaims struct {
	jwt.RegisteredClaims
	Kubernetes Kubernetes `json:"kubernetes.io"`
}

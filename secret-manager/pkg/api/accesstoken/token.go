package accesstoken

import (
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

const (
	TokenFilePath = "/var/run/secrets/secretmgr/token"
)

type AccessToken interface {
	Read() (string, error)
}

type KubernetesAccessToken struct {
	filePath     string
	data         string
	expiresAt    int64
	graceSeconds int64
	jwtParser    *jwt.Parser
}

func NewAccessToken(filePath string) AccessToken {
	return &KubernetesAccessToken{
		filePath:     filePath,
		graceSeconds: 30,
		jwtParser:    jwt.NewParser(jwt.WithExpirationRequired()),
	}
}

func (k *KubernetesAccessToken) Read() (string, error) {
	if k.data == "" || k.IsExpired() {
		if err := k.readTokenFile(); err != nil {
			return "", err
		}
	}
	return k.data, nil
}

func (k *KubernetesAccessToken) readTokenFile() error {
	if k.filePath == "" {
		return errors.New("file path is empty")
	}
	data, err := os.ReadFile(k.filePath)
	if err != nil {
		return errors.Wrapf(err, "failed to read token file %s", k.filePath)
	}
	k.data = string(data)
	if err := k.setExpiresAt(k.graceSeconds); err != nil {
		return errors.Wrap(err, "failed to set expiresAt")
	}
	if k.IsExpired() {
		return errors.New("token is expired")
	}

	return nil
}

// IsExpired checks if the token is expired
func (k *KubernetesAccessToken) IsExpired() bool {
	if k.expiresAt == 0 {
		return true
	}
	return time.Now().Unix() > k.expiresAt
}

func (k *KubernetesAccessToken) setExpiresAt(graceSeconds int64) error {
	claims, err := k.parseJwt(k.data)
	if err != nil {
		return err
	}
	k.expiresAt = claims.ExpiresAt.Unix() - graceSeconds

	return nil
}

func (k *KubernetesAccessToken) parseJwt(raw string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}
	_, _, err := k.jwtParser.ParseUnverified(raw, claims)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse token")
	}
	return claims, nil
}

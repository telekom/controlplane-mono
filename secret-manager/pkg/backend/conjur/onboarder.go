package conjur

import (
	"bytes"
	"context"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/telekom/controlplane-mono/secret-manager/pkg/backend"
	"github.com/valyala/fasttemplate"
)

var _ backend.Onboarder = &ConjurOnboarder{}

type ConjurOnboarder struct {
	conjur       ConjurAPI
	secretWriter backend.Writer[ConjurSecretId, backend.DefaultSecret[ConjurSecretId]]
	templates    map[string]*fasttemplate.Template

	bouncer Bouncer
}

func NewOnboarder(writeAPI ConjurAPI, secretWriter backend.Writer[ConjurSecretId, backend.DefaultSecret[ConjurSecretId]]) *ConjurOnboarder {
	return &ConjurOnboarder{
		conjur: writeAPI,
		templates: map[string]*fasttemplate.Template{
			"env":    fasttemplate.New(EnvironmentPolicyTemplate, startTag, endTag),
			"team":   fasttemplate.New(TeamPolicyTemplate, startTag, endTag),
			"app":    fasttemplate.New(ApplicationPolicyTemplate, startTag, endTag),
			"delete": fasttemplate.New(DeletePolicyTemplate, startTag, endTag),
		},
		secretWriter: secretWriter,
	}
}

func (c *ConjurOnboarder) WithBouncer(bouncer Bouncer) *ConjurOnboarder {
	if bouncer == nil {
		return c
	}
	c.bouncer = bouncer
	return c
}

func (c *ConjurOnboarder) OnboardEnvironment(ctx context.Context, env string) (backend.OnboardResponse, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.Info("Onboarding environment", "env", env)

	policyPath := RootPolicyPath
	buf := bytes.NewBuffer(nil)
	_, err := c.templates["env"].Execute(buf, map[string]any{
		"Environment": env,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute template")
	}

	mutator := func(ctx context.Context) error {
		log.V(1).Info("Loading policy", "policyPath", policyPath, "env", env)
		_, err = c.conjur.LoadPolicy(conjurapi.PolicyModePost, policyPath, buf)
		return err
	}

	err = c.MaybeRunWithBouncer(ctx, mutator)
	if err != nil {
		return nil, err
	}

	secretsIds, err := c.createSecrets(ctx, env, "", "", backend.EnvironmentSecrets...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create secrets for environment %s", env)
	}

	return backend.NewDefaultOnboardResponse(secretsIds), nil
}

func (c *ConjurOnboarder) OnboardTeam(ctx context.Context, env, teamId string) (backend.OnboardResponse, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.Info("Onboarding team", "env", env, "team", teamId)
	policyPath := RootPolicyPath + "/" + env

	buf := bytes.NewBuffer(nil)
	_, err := c.templates["team"].Execute(buf, map[string]any{
		"TeamId": teamId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute template")
	}

	mutator := func(ctx context.Context) error {
		log.V(1).Info("Loading policy", "policyPath", policyPath, "env", env, "teamId", teamId)
		_, err = c.conjur.LoadPolicy(conjurapi.PolicyModePost, policyPath, buf)
		return err
	}

	err = c.MaybeRunWithBouncer(ctx, mutator)
	if err != nil {
		return nil, err
	}

	secretsIds, err := c.createSecrets(ctx, env, teamId, "", backend.TeamSecrets...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create secrets for team %s", teamId)
	}

	return backend.NewDefaultOnboardResponse(secretsIds), nil
}

func (c *ConjurOnboarder) OnboardApplication(ctx context.Context, env, teamId, appId string) (backend.OnboardResponse, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.Info("Onboarding application", "env", env, "team", teamId, "app", appId)
	policyPath := RootPolicyPath + "/" + env + "/" + teamId

	buf := bytes.NewBuffer(nil)
	_, err := c.templates["app"].Execute(buf, map[string]any{
		"AppId": appId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute template")
	}

	mutator := func(ctx context.Context) error {
		log.V(1).Info("Loading policy", "policyPath", policyPath, "env", env, "teamId", teamId, "appId", appId)
		_, err = c.conjur.LoadPolicy(conjurapi.PolicyModePost, policyPath, buf)
		return err
	}

	err = c.MaybeRunWithBouncer(ctx, mutator)
	if err != nil {
		return nil, err
	}

	secretsIds, err := c.createSecrets(ctx, env, teamId, appId, backend.ApplicationSecrets...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create secrets for application %s", appId)
	}

	return backend.NewDefaultOnboardResponse(secretsIds), nil
}

func (c *ConjurOnboarder) DeleteEnvironment(ctx context.Context, env string) error {
	log := logr.FromContextOrDiscard(ctx)
	policyPath := RootPolicyPath
	log.Info("Deleting environment", "env", env, "policyPath", policyPath)

	return c.deletePolicy(ctx, policyPath, env)
}
func (c *ConjurOnboarder) DeleteTeam(ctx context.Context, env, teamId string) error {
	log := logr.FromContextOrDiscard(ctx)
	policyPath := RootPolicyPath + "/" + env
	log.Info("Deleting team", "env", env, "team", teamId, "policyPath", policyPath)
	return c.deletePolicy(ctx, policyPath, teamId)
}
func (c *ConjurOnboarder) DeleteApplication(ctx context.Context, env, teamId, appId string) error {
	log := logr.FromContextOrDiscard(ctx)
	policyPath := RootPolicyPath + "/" + env + "/" + teamId
	log.Info("Deleting application", "env", env, "team", teamId, "app", appId, "policyPath", policyPath)
	return c.deletePolicy(ctx, policyPath, appId)
}

func (c *ConjurOnboarder) deletePolicy(ctx context.Context, policyPath, policyKey string) error {
	log := logr.FromContextOrDiscard(ctx)
	buf := bytes.NewBuffer(nil)
	_, err := c.templates["delete"].Execute(buf, map[string]any{
		"PolicyPath": policyKey,
	})
	if err != nil {
		return errors.Wrap(err, "failed to execute delete template")
	}
	log.Info("Deleting policy", "policyPath", policyPath, "policyKey", policyKey)
	_, err = c.conjur.LoadPolicy(conjurapi.PolicyModePatch, policyPath, buf)
	if err != nil {
		return errors.Wrap(err, "failed to load delete policy")
	}

	return nil
}

func (c *ConjurOnboarder) createSecrets(ctx context.Context, env, teamId, appId string, secretNames ...string) (map[string]backend.SecretRef, error) {
	log := logr.FromContextOrDiscard(ctx)
	secretsIds := make(map[string]backend.SecretRef)
	if c.secretWriter == nil {
		return secretsIds, nil
	}
	for _, secretName := range secretNames {
		secretId := New(env, teamId, appId, secretName, "")
		log.Info("Creating secret", "secretId", secretId.String())
		secret, err := c.secretWriter.Set(ctx, secretId, backend.InitialString(uuid.NewString()))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to initialize secret %s", secretId.VariableId())
		}
		secretsIds[secretName] = secret.Id()
	}

	return secretsIds, nil
}

func (c *ConjurOnboarder) MaybeRunWithBouncer(ctx context.Context, runnable Runnable) error {
	if c.bouncer == nil {
		return runnable(ctx)
	}
	err := <-c.bouncer.Run(ctx, runnable)
	if err != nil && errors.Is(err, ErrQueueFull) {
		return backend.NewBackendError(nil, err, backend.TypeErrTooManyRequests)
	}
	return err
}

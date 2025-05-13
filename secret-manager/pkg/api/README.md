# Secret Manager API

## Overview

The Secret Manager API provides a way to manage secrets in a secure and efficient manner. It allows you to get and set secrets, as well as list all secrets. The API is designed to be easy to use and integrate with other services.


## Components

This module offers the following components:

### Secrets API

This API allows you to manage secrets. You can get and set secrets. The API is designed to be easy to use and integrate with other services.

```go
// Global default API
api.Get(ctx, "{{poc:eni--hyperion:my-foo-app:clientSecret:<some-hash>}}")
api.Set(ctx, "{{poc:eni--hyperion:my-foo-app:clientSecret:<some-hash>}}", "my-new-value")

// API with custom options
// Note: This API needs the clean ID of the secret without the start and end tags
secretsApi := api.NewSecrets()
secretsApi.Set(ctx, "poc:eni--hyperion:my-foo-app:clientSecret:<some-hash>", "my-new-value")
secretsApi.Get(ctx, "poc:eni--hyperion:my-foo-app:clientSecret:<some-hash>")
```

The global API is automatically initialized with the default options. It is recommended to use the global API for most use cases.
It will detect if the service is running in a local or Kubernetes environment and use the appropriate configuration.

### Onboarding API

This API allows you to manage the onboarding process. You can create and delete organizational structures like `Environments`, `Teams` and `Applications`.
This API should only be used by special services that are responsible for onboarding new customers. It is not intended to be used by regular users or applications.

```go
onboardingApi := api.NewOnboarding()

// Create a new environment
onboardingApi.UpsertEnvironment(ctx, "poc")

// Create a new team
onboardingApi.UpsertTeam(ctx, "poc", "eni--hyperion")

// Create a new application
onboardingApi.UpsertApplication(ctx, "poc", "eni--hyperion", "my-foo-app")
```

## Vocabulary

| Name | Description |
| ---- | ----------- |
| `environment` | The environment is the top level organizational structure. |
| `team` | The team is the second level organizational structure. It is dynamically configured per onboarded team. |
| `application` | The application is the third level organizational structure. It is dynamically configured per onboarded application of a team. |
|------------|-----------------|
| `secret` | The secret is the actual secret value. It can be owned by any structurical element. |
| `secretId` | The secret ID is the unique identifier of the secret. It is used to reference the secret in the API. |
| `secretName` | The secret name that is part of the secret ID. It is used to identity the secret in the context of the structural element.
| `secretPlaceholder` or `secretRef` | The secret placeholder is a variation of the secret ID using prefix and suffix tags. It is used to immidiately identify the secret. |
|------------|-----------------|
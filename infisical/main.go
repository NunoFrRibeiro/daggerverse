// Module to interact with an Infisical project
// (for now this module only retrieves secrets from a project authenticating using Universal Auth)
package main

import (
	"context"
	"dagger/infisical/internal/dagger"

	infisical "github.com/infisical/go-sdk"
)

type Infisical struct {
	// +private
	SiteUrl string
	// +private
	UniversalAuthClientID *dagger.Secret
	// +private
	UniversalAuthClientSecret *dagger.Secret
}

func New(
	// The URL of the Infisical API. Default is
	// +optional
	// default="https://api.infisical.com"
	siteUrl string,

	// Your machine identity client ID
	// +required
	universalAuthClientID *dagger.Secret,

	// Your machine identity client secret
	// +required
	universalAuthClientSecret *dagger.Secret,
) *Infisical {
	return &Infisical{
		SiteUrl:                   siteUrl,
		UniversalAuthClientID:     universalAuthClientID,
		UniversalAuthClientSecret: universalAuthClientSecret,
	}
}

func (m *Infisical) GetSecret(
	ctx context.Context,

	// The key of the secret to retrieve
	// +required
	secretKey string,

	// The project ID where the secret lives in
	// +required
	projectId string,

	// The slug name (dev, prod, etc) of the environment from where secrets should be fetched from
	// +required
	environment string,

	// The path from where secret should be fetched from
	// +optional
	secretPath string,

	// The type of the secret. Valid options are “shared” or “personal”. If not specified, the default value is “shared”
	// +optional
	secretType string,
) (*dagger.Secret, error) {

	// create an instance
	infisicalClient := infisical.NewInfisicalClient(infisical.Config{
		SiteUrl: m.SiteUrl,
	})

	clientID, err := m.UniversalAuthClientID.Plaintext(ctx)
	if err != nil {
		return nil, err
	}
	clientSecret, err := m.UniversalAuthClientSecret.Plaintext(ctx)
	if err != nil {
		return nil, err
	}

	_, err = infisicalClient.Auth().
		UniversalAuthLogin(clientID, clientSecret)
	if err != nil {
		return nil, err
	}

	secret, err := infisicalClient.
		Secrets().
		Retrieve(infisical.RetrieveSecretOptions{
			SecretKey:   secretKey,
			ProjectID:   projectId,
			Environment: environment,
			SecretPath:  secretPath,
			Type:        secretType,
		})

	if err != nil {
		return nil, err
	}

	returnValue := secret.SecretValue

	return dag.SetSecret("val", returnValue), nil
}

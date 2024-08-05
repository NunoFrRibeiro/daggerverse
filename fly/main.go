// Module for interacting with Fly.io apps
// (For the moment it's used to deploy to an existing app)
package main

import (
	"dagger/flyio/internal/dagger"
	"fmt"

	"golang.org/x/net/context"
)

// If no version of fly is specified, it defaults to latest
const flyVersion = "latest"

type Flyio struct {
	// +private
	Container *dagger.Container
	// +private
	Version string
	// +private
	Regions string
	// +private
	Org string
}

func New(
	// Fly.io Container version
	// +optional
	version string,

	// Token to connect to fly.io,
	token *dagger.Secret,

	// Deploy to machines only in these regions.
	// Multiple regions can be specified with comma separated valuesor
	// or by providing the flag multiple times
	// Ex: mad,fra
	// +optional
	regions string,

	// The target Fly.io organizatio (defaults to personal)
	// +optional
	// +default="personal"
	org string,

	// Use a custom container with Fly installed
	// +optional
	container *dagger.Container,

) *Flyio {
	if container == nil {
		if version == "" {
			version = flyVersion
		}
		container = dag.Container().
			From(fmt.Sprintf("flyio/flyctl:%s", version)).
			WithSecretVariable("FLY_API_TOKEN", token)
	}
	return &Flyio{
		Container: container,
		Version:   version,
		Regions:   regions,
		Org:       org,
	}
}

func (m *Flyio) Deploy(
	ctx context.Context,
	// Directory from where to deploy the app (assumes there's a fly.toml in the directory)
	dir *dagger.Directory,

	// Deploy using a Docker image
	// +optional
	image string,
) (string, error) {
	deployCommand := []string{
		"deploy",
		"--regions",
		m.Regions,
	}

	if image != "" {
		deployCommand = []string{
			"deploy",
			"--regions",
			m.Regions,
			"--image",
			image,
		}
	}
	return m.Container.
		WithMountedDirectory("/app", dir).
		WithWorkdir("/app").
		WithExec(deployCommand, dagger.ContainerWithExecOpts{
			UseEntrypoint: true,
		}).Stdout(ctx)
}

// Create a new application on the Fly platform
func (m *Flyio) Create(
	ctx context.Context,
	// App name
	// Ex: `goblog-2024-08-06`
	appName string,
) (string, error) {
	return m.Container.
		WithExec([]string{
			"create",
			"--name",
			appName,
			"-o",
			m.Org,
		}, dagger.ContainerWithExecOpts{
			UseEntrypoint: true,
		}).Stdout(ctx)
}

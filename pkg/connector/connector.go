package connector

import (
	"context"
	"fmt"
	"io"

	"github.com/buildkite/go-buildkite/v4"
	"github.com/conductorone/baton-buildkite/pkg/config"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

type Connector struct {
	client *buildkite.Client
	org    string
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncerV2 {
	return []connectorbuilder.ResourceSyncerV2{
		newUserBuilder(d.client, d.org),
		newTeamBuilder(d.client, d.org),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Buildkite",
		Description: "Connector syncing users and teams from Buildkite.",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	_, _, err := d.client.Organizations.Get(ctx, d.org)
	if err != nil {
		return nil, fmt.Errorf("baton-buildkite: failed to validate credentials for org %q: %w", d.org, err)
	}
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, cfg *config.Buildkite, cliOpts *cli.ConnectorOpts) (connectorbuilder.ConnectorBuilderV2, []connectorbuilder.Opt, error) {
	c, err := buildkite.NewOpts(buildkite.WithTokenAuth(cfg.ApiToken))
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to create client: %w", err)
	}

	return &Connector{
		client: c,
		org:    cfg.Organization,
	}, nil, nil
}

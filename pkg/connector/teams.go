package connector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/buildkite/go-buildkite/v4"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type teamBuilder struct {
	client *buildkite.Client
	org    string
}

func (t *teamBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return teamResourceType
}

// List returns all the teams as resource objects.
func (t *teamBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts resourceSdk.SyncOpAttrs) ([]*v2.Resource, *resourceSdk.SyncOpResults, error) {
	page, err := pageTokenToInt(opts.PageToken.Token)
	if err != nil {
		return nil, nil, err
	}

	teams, resp, err := t.client.Teams.List(ctx, t.org, &buildkite.TeamsListOptions{
		ListOptions: buildkite.ListOptions{
			Page: page,
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to list teams: %w", err)
	}

	resources := make([]*v2.Resource, 0, len(teams))
	for _, team := range teams {
		resource, err := resourceSdk.NewGroupResource(
			team.Name,
			teamResourceType,
			team.ID,
			[]resourceSdk.GroupTraitOption{},
		)
		if err != nil {
			return nil, nil, fmt.Errorf("baton-buildkite: failed to create team resource: %w", err)
		}
		resources = append(resources, resource)
	}

	nextPageToken := strconv.Itoa(resp.NextPage)
	if resp.NextPage == 0 {
		nextPageToken = ""
	}

	return resources, &resourceSdk.SyncOpResults{
		NextPageToken: nextPageToken,
	}, nil
}

// Entitlements always returns an empty slice for teams.
func (t *teamBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

// Grants always returns an empty slice for teams.
func (t *teamBuilder) Grants(ctx context.Context, resource *v2.Resource, opts resourceSdk.SyncOpAttrs) ([]*v2.Grant, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

func newTeamBuilder(client *buildkite.Client, org string) *teamBuilder {
	return &teamBuilder{
		client: client,
		org:    org,
	}
}

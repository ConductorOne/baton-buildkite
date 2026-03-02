package connector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/buildkite/go-buildkite/v4"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type organizationBuilder struct {
	client *buildkite.Client
}

func (o *organizationBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return organizationResourceType
}

// List returns all the organizations as resource objects.
func (o *organizationBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts resourceSdk.SyncOpAttrs) ([]*v2.Resource, *resourceSdk.SyncOpResults, error) {
	page, err := pageTokenToInt(opts.PageToken.Token)
	if err != nil {
		return nil, nil, err
	}

	orgs, resp, err := o.client.Organizations.List(ctx, &buildkite.OrganizationListOptions{
		ListOptions: buildkite.ListOptions{
			Page: page,
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to list organizations: %w", err)
	}

	resources := make([]*v2.Resource, 0, len(orgs))
	for _, org := range orgs {
		resource, err := resourceSdk.NewResource(
			org.Name,
			organizationResourceType,
			org.ID,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("baton-buildkite: failed to create organization resource: %w", err)
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

// Entitlements always returns an empty slice for organizations.
func (o *organizationBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

// Grants always returns an empty slice for organizations.
func (o *organizationBuilder) Grants(ctx context.Context, resource *v2.Resource, opts resourceSdk.SyncOpAttrs) ([]*v2.Grant, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

func newOrganizationBuilder(client *buildkite.Client) *organizationBuilder {
	return &organizationBuilder{
		client: client,
	}
}

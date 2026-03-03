package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-buildkite/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userBuilder struct {
	client *client.Client
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts resourceSdk.SyncOpAttrs) ([]*v2.Resource, *resourceSdk.SyncOpResults, error) {
	var bag pagination.Bag
	err := bag.Unmarshal(opts.PageToken.Token)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to unmarshal page token: %w", err)
	}

	if bag.Current() == nil {
		bag.Push(pagination.PageState{
			ResourceTypeID: userResourceType.Id,
		})
	}

	members, nextPageURL, err := o.client.ListMembers(ctx, bag.PageToken())
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to list members: %w", err)
	}

	var resources []*v2.Resource
	for _, m := range members {
		r, err := resourceSdk.NewUserResource(
			m.Name,
			userResourceType,
			m.ID,
			[]resourceSdk.UserTraitOption{
				resourceSdk.WithEmail(m.Email, true),
				resourceSdk.WithStatus(v2.UserTrait_Status_STATUS_ENABLED),
			},
		)
		if err != nil {
			return nil, nil, fmt.Errorf("baton-buildkite: failed to create user resource: %w", err)
		}
		resources = append(resources, r)
	}

	nextPage, err := bag.NextToken(nextPageURL)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to create next page token: %w", err)
	}

	return resources, &resourceSdk.SyncOpResults{NextPageToken: nextPage}, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, opts resourceSdk.SyncOpAttrs) ([]*v2.Grant, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

func newUserBuilder(c *client.Client) *userBuilder {
	return &userBuilder{client: c}
}

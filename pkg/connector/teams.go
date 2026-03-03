package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-buildkite/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	sdkEntitlement "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	sdkGrant "github.com/conductorone/baton-sdk/pkg/types/grant"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type teamBuilder struct {
	client *client.Client
}

func (t *teamBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return teamResourceType
}

func (t *teamBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts resourceSdk.SyncOpAttrs) ([]*v2.Resource, *resourceSdk.SyncOpResults, error) {
	var bag pagination.Bag
	err := bag.Unmarshal(opts.PageToken.Token)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to unmarshal page token: %w", err)
	}

	if bag.Current() == nil {
		bag.Push(pagination.PageState{
			ResourceTypeID: teamResourceType.Id,
		})
	}

	teams, nextPageURL, err := t.client.ListTeams(ctx, bag.PageToken())
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to list teams: %w", err)
	}

	var resources []*v2.Resource
	for _, team := range teams {
		profile := map[string]interface{}{
			"description":     team.Description,
			"privacy":         team.Privacy,
			"slug":            team.Slug,
			"is_default_team": team.IsDefaultTeam,
		}

		r, err := resourceSdk.NewGroupResource(
			team.Name,
			teamResourceType,
			team.ID,
			[]resourceSdk.GroupTraitOption{
				resourceSdk.WithGroupProfile(profile),
			},
		)
		if err != nil {
			return nil, nil, fmt.Errorf("baton-buildkite: failed to create team resource: %w", err)
		}
		resources = append(resources, r)
	}

	nextPage, err := bag.NextToken(nextPageURL)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to create next page token: %w", err)
	}

	return resources, &resourceSdk.SyncOpResults{NextPageToken: nextPage}, nil
}

func (t *teamBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
	entitlements := []*v2.Entitlement{
		sdkEntitlement.NewAssignmentEntitlement(
			resource,
			"member",
			sdkEntitlement.WithDisplayName(fmt.Sprintf("%s Team Member", resource.DisplayName)),
			sdkEntitlement.WithDescription(fmt.Sprintf("Member of the %s team in Buildkite", resource.DisplayName)),
			sdkEntitlement.WithGrantableTo(userResourceType),
		),
	}

	return entitlements, nil, nil
}

func (t *teamBuilder) Grants(ctx context.Context, resource *v2.Resource, opts resourceSdk.SyncOpAttrs) ([]*v2.Grant, *resourceSdk.SyncOpResults, error) {
	var bag pagination.Bag
	err := bag.Unmarshal(opts.PageToken.Token)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to unmarshal page token: %w", err)
	}

	if bag.Current() == nil {
		bag.Push(pagination.PageState{
			ResourceTypeID: teamResourceType.Id,
			ResourceID:     resource.Id.Resource,
		})
	}

	teamMembers, nextPageURL, err := t.client.ListTeamMembers(ctx, resource.Id.Resource, bag.PageToken())
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to list team members: %w", err)
	}

	var grants []*v2.Grant
	for _, m := range teamMembers {
		principalID := &v2.ResourceId{
			ResourceType: userResourceType.Id,
			Resource:     m.UserID,
		}
		grants = append(grants, sdkGrant.NewGrant(resource, "member", principalID))
	}

	nextPage, err := bag.NextToken(nextPageURL)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to create next page token: %w", err)
	}

	return grants, &resourceSdk.SyncOpResults{NextPageToken: nextPage}, nil
}

func newTeamBuilder(c *client.Client) *teamBuilder {
	return &teamBuilder{client: c}
}

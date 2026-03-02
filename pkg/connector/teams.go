package connector

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/buildkite/go-buildkite/v4"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"google.golang.org/grpc/codes"
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
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			return nil, nil, uhttp.WrapErrorsWithRateLimitInfo(codes.Unavailable, resp.Response, err)
		}
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

// Entitlements returns the member and maintainer entitlements for the team.
func (t *teamBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
	return []*v2.Entitlement{
		entitlement.NewAssignmentEntitlement(
			resource,
			"member",
			entitlement.WithDisplayName("Member"),
			entitlement.WithDescription(fmt.Sprintf("Member of %s team in Buildkite", resource.DisplayName)),
			entitlement.WithGrantableTo(userResourceType),
		),
		entitlement.NewAssignmentEntitlement(
			resource,
			"maintainer",
			entitlement.WithDisplayName("Maintainer"),
			entitlement.WithDescription(fmt.Sprintf("Maintainer of %s team in Buildkite", resource.DisplayName)),
			entitlement.WithGrantableTo(userResourceType),
		),
	}, &resourceSdk.SyncOpResults{}, nil
}

// Grants returns all users who are members of this team with their roles.
func (t *teamBuilder) Grants(ctx context.Context, resource *v2.Resource, opts resourceSdk.SyncOpAttrs) ([]*v2.Grant, *resourceSdk.SyncOpResults, error) {
	teamID := resource.Id.Resource

	page, err := pageTokenToInt(opts.PageToken.Token)
	if err != nil {
		return nil, nil, err
	}

	members, resp, err := t.client.TeamMember.ListTeamMembers(ctx, t.org, teamID, &buildkite.TeamMembersListOptions{
		ListOptions: buildkite.ListOptions{
			Page: page,
		},
	})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			return nil, nil, uhttp.WrapErrorsWithRateLimitInfo(codes.Unavailable,
				resp.Response,
				fmt.Errorf("baton-buildkite: failed to list team members: %w", err),
			)
		}
		return nil, nil, fmt.Errorf("baton-buildkite: failed to list team members: %w", err)
	}

	grants := make([]*v2.Grant, 0, len(members))
	for _, member := range members {
		grants = append(grants, grant.NewGrant(
			resource,
			member.Role,
			&v2.ResourceId{
				ResourceType: userResourceType.Id,
				Resource:     member.ID,
			},
		))
	}

	nextPageToken := strconv.Itoa(resp.NextPage)
	if resp.NextPage == 0 {
		nextPageToken = ""
	}

	return grants, &resourceSdk.SyncOpResults{
		NextPageToken: nextPageToken,
	}, nil
}

func newTeamBuilder(client *buildkite.Client, org string) *teamBuilder {
	return &teamBuilder{
		client: client,
		org:    org,
	}
}

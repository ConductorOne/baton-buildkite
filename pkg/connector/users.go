package connector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/buildkite/go-buildkite/v4"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userBuilder struct {
	client *buildkite.Client
	org    string
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts resourceSdk.SyncOpAttrs) ([]*v2.Resource, *resourceSdk.SyncOpResults, error) {
	page, err := pageTokenToInt(opts.PageToken.Token)
	if err != nil {
		return nil, nil, err
	}

	bMembers, resp, err := o.client.Members.List(ctx, o.org, &buildkite.MemberListOptions{
		ListOptions: buildkite.ListOptions{
			Page: page,
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to list users: %w", err)
	}

	users := make([]*v2.Resource, 0, len(bMembers))
	for _, member := range bMembers {
		if err := ctx.Err(); err != nil {
			return nil, nil, fmt.Errorf("baton-buildkite: users sync canceled: %w", err)
		}

		userResource, err := resourceSdk.NewUserResource(
			member.Name,
			userResourceType,
			member.UUID,
			[]resourceSdk.UserTraitOption{
				resourceSdk.WithEmail(member.Email, true),
				resourceSdk.WithUserLogin(member.Email),
				resourceSdk.WithStatus(v2.UserTrait_Status_STATUS_ENABLED),
			},
			resourceSdk.WithExternalID(&v2.ExternalId{Id: member.UUID}),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("baton-buildkite: failed to create user resource: %w", err)
		}
		users = append(users, userResource)
	}

	nextPageToken := ""
	if resp.NextPage != 0 {
		nextPageToken = strconv.Itoa(resp.NextPage)
	}

	return users, &resourceSdk.SyncOpResults{
		NextPageToken: nextPageToken,
	}, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, opts resourceSdk.SyncOpAttrs) ([]*v2.Grant, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

func newUserBuilder(client *buildkite.Client, org string) *userBuilder {
	return &userBuilder{
		client: client,
		org:    org,
	}
}

package connector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/buildkite/go-buildkite/v4"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
)

const listUsersURL = "v2/organizations/%s/members"

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

	// go-buildkite doesn't have a built-in method for org members yet.
	// See: https://github.com/buildkite/go-buildkite/pull/282
	u := fmt.Sprintf(listUsersURL, o.org)
	u, err = addOptions(u, buildkite.ListOptions{
		Page: page,
	})
	if err != nil {
		return nil, nil, err
	}

	req, err := o.client.NewRequest(ctx, "GET", u, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to create request: %w", err)
	}

	var bUsers []*buildkite.User
	resp, err := o.client.Do(req, &bUsers)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-buildkite: failed to list users: %w", err)
	}

	users := make([]*v2.Resource, 0, len(bUsers))
	for _, user := range bUsers {
		userResource, err := resourceSdk.NewUserResource(
			user.Name,
			userResourceType,
			user.ID,
			[]resourceSdk.UserTraitOption{
				resourceSdk.WithEmail(user.Email, true),
				resourceSdk.WithUserLogin(user.Email),
				resourceSdk.WithStatus(v2.UserTrait_Status_STATUS_ENABLED),
			},
			resourceSdk.WithExternalID(&v2.ExternalId{Id: user.ID}),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("baton-buildkite: failed to create user resource: %w", err)
		}
		users = append(users, userResource)
	}

	if resp == nil {
		return users, nil, nil
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

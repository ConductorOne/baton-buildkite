package connector

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strconv"

	"github.com/buildkite/go-buildkite/v4"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/google/go-querystring/query"
)

var listUsersUrl = "v2/organizations/%s/members"

type userBuilder struct {
	client *buildkite.Client
	org    string
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts resourceSdk.SyncOpAttrs) ([]*v2.Resource, *resourceSdk.SyncOpResults, error) {
	page, err := pageTokenToInt(opts.PageToken.Token)
	if err != nil {
		return nil, nil, err
	}

	// TODO. (Ben.Su) https://github.com/buildkite/go-buildkite/pull/282.
	// We should use this built-in function once this feature is in.
	u := fmt.Sprintf(listUsersUrl, o.org)
	u, err = addOptions(u, buildkite.ListOptions{
		Page: page,
	})
	if err != nil {
		return nil, nil, err
	}

	req, err := o.client.NewRequest(ctx, "GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var bUsers []*buildkite.User
	resp, err := o.client.Do(req, &bUsers)
	if err != nil {
		return nil, nil, err
	}

	users := make([]*v2.Resource, 0, len(bUsers))
	for _, user := range bUsers {
		userResource, err := resourceSdk.NewUserResource(
			user.Name,
			userResourceType,
			user.ID,
			[]resourceSdk.UserTraitOption{resourceSdk.WithEmail(user.Email, true)},
		)
		if err != nil {
			return nil, nil, fmt.Errorf("baton-buildkite: failed to create user resource: %w", err)
		}
		users = append(users, userResource)
	}

	nextPageToken := strconv.Itoa(resp.NextPage)
	if resp.NextPage == 0 {
		nextPageToken = ""
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

// addOptions adds the parameters in opt as URL query parameters to s.  opt
// must be a struct whose fields may contain "url" tags.
func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

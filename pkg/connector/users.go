package connector

import (
	"context"

	"github.com/conductorone/baton-buildkite/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userBuilder struct {
	c   client.Client
	org string
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

func userResource(user *client.User, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	}

	userTraitOptions := []resourceSdk.UserTraitOption{
		resourceSdk.WithEmail(user.Email, true),
		resourceSdk.WithStatus(v2.UserTrait_Status_STATUS_ENABLED),
		resourceSdk.WithUserProfile(profile),
	}

	return resourceSdk.NewUserResource(
		user.Name,
		userResourceType,
		user.ID,
		userTraitOptions,
		resourceSdk.WithParentResourceID(parentResourceID),
	)
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	outputAnnotations := annotations.Annotations{}
	resp, err := o.c.ListUsers(ctx, o.org, pToken.Token)
	if resp != nil {
		outputAnnotations.WithRateLimiting(resp.RateLimitDescription)
	}
	if err != nil {
		return nil, "", outputAnnotations, err
	}
	outputResources := make([]*v2.Resource, 0, len(resp.Users))
	for _, user := range resp.Users {
		userResource0, err := userResource(user, parentResourceID)
		if err != nil {
			return nil, "", nil, err
		}
		outputResources = append(outputResources, userResource0)
	}

	return outputResources, resp.Pagination, outputAnnotations, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(c client.Client, org string) *userBuilder {
	return &userBuilder{
		c:   c,
		org: org,
	}
}

package connector

import (
	"context"

	"github.com/conductorone/baton-dockerhub/pkg/dockerhub"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userResourceType struct {
	resourceType *v2.ResourceType
	client       *dockerhub.Client
}

func (u *userResourceType) ResourceType(ctx context.Context) *v2.ResourceType {
	return resourceTypeUser
}

// Create a new connector resource for an DockerHub user.
func userResource(ctx context.Context, user *dockerhub.User, parentId *v2.ResourceId) (*v2.Resource, error) {
	firstName, lastName := splitFullName(user.FullName)

	profile := map[string]interface{}{
		"first_name": firstName,
		"last_name":  lastName,
		"login":      user.Username,
		"user_id":    user.Id,
	}

	userTraitOptions := []rs.UserTraitOption{
		rs.WithUserProfile(profile),
		rs.WithStatus(v2.UserTrait_Status_STATUS_ENABLED),
		rs.WithEmail(user.Email, true),
	}

	resource, err := rs.NewUserResource(
		user.Username,
		resourceTypeUser,
		user.Id,
		userTraitOptions,
		rs.WithParentResourceID(parentId),
	)

	if err != nil {
		return nil, err
	}

	return resource, nil
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (u *userResourceType) List(ctx context.Context, parentId *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentId == nil {
		return nil, "", nil, nil
	}

	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: resourceTypeUser.Id})
	if err != nil {
		return nil, "", nil, err
	}

	paginationOpts := dockerhub.PaginationVars{
		Size: ResourcesPageSize,
		Page: page,
	}

	users, nextPage, err := u.client.ListUsers(ctx, parentId.Resource, &paginationOpts)
	if err != nil {
		return nil, "", nil, err
	}

	next, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Resource
	for _, user := range users {
		userCopy := user

		ur, err := userResource(ctx, &userCopy, parentId)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, ur)
	}

	return rv, next, nil, nil
}

// Entitlements always returns an empty slice for users.
func (u *userResourceType) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (u *userResourceType) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func userBuilder(client *dockerhub.Client) *userResourceType {
	return &userResourceType{
		resourceType: resourceTypeUser,
		client:       client,
	}
}

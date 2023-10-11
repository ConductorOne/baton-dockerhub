package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-dockerhub/pkg/dockerhub"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

const (
	readPermission         = "read"
	readAndWritePermission = "write"
	adminPermission        = "admin"
)

var repoPermissions = []string{readPermission, readAndWritePermission, adminPermission}

type repositoryResourceType struct {
	resourceType *v2.ResourceType
	client       *dockerhub.Client
}

func (r *repositoryResourceType) ResourceType(ctx context.Context) *v2.ResourceType {
	return resourceTypeRepository
}

// Create a new connector resource for an DockerHub repository.
func repositoryResource(ctx context.Context, repository *dockerhub.Repository, parentId *v2.ResourceId) (*v2.Resource, error) {
	resource, err := rs.NewResource(
		titleCase(repository.Name),
		resourceTypeRepository,
		repository.Name,
		rs.WithParentResourceID(parentId),
		rs.WithDescription(repository.Description),
	)

	if err != nil {
		return nil, err
	}

	return resource, nil
}

// List returns all the repositories from the database as resource objects.
func (r *repositoryResourceType) List(ctx context.Context, parentId *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentId == nil {
		return nil, "", nil, nil
	}

	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: resourceTypeRepository.Id})
	if err != nil {
		return nil, "", nil, err
	}

	paginationOpts := dockerhub.PaginationVars{
		Size: ResourcesPageSize,
		Page: page,
	}

	repositories, nextPage, err := r.client.ListRepositories(ctx, parentId.Resource, &paginationOpts)
	if err != nil {
		return nil, "", nil, fmt.Errorf("dockerhub-connector: failed to list repositories: %w", err)
	}

	next, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Resource
	for _, repository := range repositories {
		repositoryCopy := repository

		rr, err := repositoryResource(ctx, &repositoryCopy, parentId)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, rr)
	}

	return rv, next, nil, nil
}

// Entitlements returns a slice of entitlements for possible permissions of repositories (read, read & write, admin).
func (r *repositoryResourceType) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	for _, p := range repoPermissions {
		permissionOptions := []ent.EntitlementOption{
			ent.WithGrantableTo(resourceTypeTeam),
			ent.WithDisplayName(fmt.Sprintf("%s Repository %s", resource.DisplayName, p)),
			ent.WithDescription(fmt.Sprintf("%s access to %s repository in DockerHub", titleCase(p), resource.DisplayName)),
		}

		rv = append(rv, ent.NewPermissionEntitlement(resource, p, permissionOptions...))
	}

	return rv, "", nil, nil
}

// Grants returns a slice of grants for each team permission set in repositories.
func (r *repositoryResourceType) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	repoId, orgSlug := resource.Id.Resource, resource.ParentResourceId.Resource
	bag, page, err := parsePageToken(pToken.Token, resource.Id)
	if err != nil {
		return nil, "", nil, err
	}

	paginationOpts := dockerhub.PaginationVars{
		Size: ResourcesPageSize,
		Page: page,
	}

	perms, nextPage, err := r.client.ListRepositoryPermissions(ctx, orgSlug, repoId, &paginationOpts)
	if err != nil {
		return nil, "", nil, fmt.Errorf("dockerhub-connector: failed to list repository permissions: %w", err)
	}

	next, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Grant
	for _, perm := range perms {
		// fetch team Id from obtained team name
		team, err := r.client.GetTeam(ctx, orgSlug, perm.TeamName)
		if err != nil {
			return nil, "", nil, fmt.Errorf("dockerhub-connector: failed to get team: %w", err)
		}

		g := grant.NewGrant(
			resource,
			perm.Permission,
			&v2.ResourceId{
				ResourceType: resourceTypeTeam.Id,
				Resource:     fmt.Sprintf("%d", team.Id),
			},
			grant.WithAnnotation(
				&v2.GrantExpandable{
					EntitlementIds:  []string{fmt.Sprintf("team:%d:%s", team.Id, teamMembership)},
					Shallow:         true,
					ResourceTypeIds: []string{resourceTypeUser.Id},
				},
			),
		)

		rv = append(rv, g)
	}

	return rv, next, nil, nil
}

func repositoryBuilder(client *dockerhub.Client) *repositoryResourceType {
	return &repositoryResourceType{
		resourceType: resourceTypeRepository,
		client:       client,
	}
}

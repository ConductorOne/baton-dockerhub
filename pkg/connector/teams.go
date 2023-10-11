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
	teamMembership = "member"
)

type teamResourceType struct {
	resourceType *v2.ResourceType
	client       *dockerhub.Client
}

func (t *teamResourceType) ResourceType(ctx context.Context) *v2.ResourceType {
	return resourceTypeTeam
}

// Create a new connector resource for an DockerHub team.
func teamResource(ctx context.Context, team *dockerhub.Team, parentId *v2.ResourceId) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"team_id":   team.Id,
		"team_name": team.Name,
	}

	teamTraitOptions := []rs.GroupTraitOption{
		rs.WithGroupProfile(profile),
	}

	resource, err := rs.NewGroupResource(
		team.Name,
		resourceTypeTeam,
		team.Id,
		teamTraitOptions,
		rs.WithParentResourceID(parentId),
		rs.WithDescription(team.Description),
	)

	if err != nil {
		return nil, err
	}

	return resource, nil
}

// List returns all the teams from the database as resource objects.
// Teams include a GroupTrait because they are the 'shape' of a standard group.
func (t *teamResourceType) List(ctx context.Context, parentId *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentId == nil {
		return nil, "", nil, nil
	}

	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: resourceTypeTeam.Id})
	if err != nil {
		return nil, "", nil, err
	}

	paginationOpts := dockerhub.PaginationVars{
		Size: ResourcesPageSize,
		Page: page,
	}

	teams, nextPage, err := t.client.ListTeams(ctx, parentId.Resource, &paginationOpts)
	if err != nil {
		return nil, "", nil, fmt.Errorf("dockerhub-connector: failed to list teams: %w", err)
	}

	next, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Resource
	for _, team := range teams {
		teamCopy := team

		tr, err := teamResource(ctx, &teamCopy, parentId)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, tr)
	}

	return rv, next, nil, nil
}

// Entitlements returns always one membership entitlement representing that a user is a member of a team.
func (t *teamResourceType) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	assignmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(resourceTypeUser),
		ent.WithDisplayName(fmt.Sprintf("%s Team %s", resource.DisplayName, teamMembership)),
		ent.WithDescription(fmt.Sprintf("Access to %s team in DockerHub", resource.DisplayName)),
	}

	// create membership entitlement
	rv = append(rv, ent.NewAssignmentEntitlement(
		resource,
		teamMembership,
		assignmentOptions...,
	))

	return rv, "", nil, nil
}

// Grants returns a slice of grants for each member that team contain.
func (t *teamResourceType) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	teamGroupTrait, err := rs.GetGroupTrait(resource)
	if err != nil {
		return nil, "", nil, err
	}

	teamSlug, ok := rs.GetProfileStringValue(teamGroupTrait.Profile, "team_name")
	if !ok {
		return nil, "", nil, fmt.Errorf("dockerhub-connector: failed to get team name from profile")
	}

	orgSlug := resource.ParentResourceId.Resource
	bag, page, err := parsePageToken(pToken.Token, resource.Id)
	if err != nil {
		return nil, "", nil, err
	}

	paginationOpts := dockerhub.PaginationVars{
		Size: ResourcesPageSize,
		Page: page,
	}

	members, nextPage, err := t.client.ListTeamMembers(ctx, orgSlug, teamSlug, &paginationOpts)
	if err != nil {
		return nil, "", nil, fmt.Errorf("dockerhub-connector: failed to list team members: %w", err)
	}

	next, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Grant
	for _, member := range members {
		memberCopy := member
		ur, err := userResource(ctx, &memberCopy, resource.Id)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, grant.NewGrant(resource, teamMembership, ur.Id))
	}

	return rv, next, nil, nil
}

func teamBuilder(client *dockerhub.Client) *teamResourceType {
	return &teamResourceType{
		resourceType: resourceTypeTeam,
		client:       client,
	}
}

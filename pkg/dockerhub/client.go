package dockerhub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-dockerhub/pkg/dockerhub/external"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	BaseDomain = "hub.docker.com"

	OrgsEndpoint      = "/v2/orgs"
	OrgDetailEndpoint = OrgsEndpoint + "/%s"
	UsersEndpoint     = OrgsEndpoint + "/%s/members"

	CurrentUserEndpoint = "/v2/user"
	UserEndpoint        = "/v2/users/%s"
	UserOrgsEndpoint    = UserEndpoint + "/orgs"

	TeamsEndpoint           = OrgsEndpoint + "/%s/groups"
	TeamDetailEndpoint      = TeamsEndpoint + "/%s"
	TeamMembersEndpoint     = TeamDetailEndpoint + "/members"
	TeamPermissionsEndpoint = TeamDetailEndpoint + "/repositories"

	RepositoriesEndpoint  = "/v2/repositories/%s"
	RepositoryPermissions = RepositoriesEndpoint + "/%s/groups"
)

type Client struct {
	httpClient  *http.Client
	baseUrl     *url.URL
	currentUser string

	username     string
	password     string
	token        string
	refreshToken string
}

func NewClient(ctx context.Context, httpClient *http.Client, username, password string) (*Client, error) {
	token, refreshToken, err := external.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}

	base := &url.URL{
		Scheme: "https",
		Host:   BaseDomain,
	}

	return &Client{
		httpClient:   httpClient,
		baseUrl:      base,
		username:     username,
		password:     password,
		token:        token,
		refreshToken: refreshToken,
	}, nil
}

func (c *Client) composeURL(endpoint string, params ...interface{}) *url.URL {
	return c.baseUrl.ResolveReference(
		&url.URL{
			Path: fmt.Sprintf(endpoint, params...),
		},
	)
}

type PaginationVars struct {
	Size uint
	Page string
}

type ListResponse[T any] struct {
	PaginationData
	Results []T `json:"results"`
}

// SetCurrentUser sets the current user for the client.
func (c *Client) SetCurrentUser(ctx context.Context) error {
	var response User

	err := c.doRequest(
		ctx,
		http.MethodGet,
		c.composeURL(CurrentUserEndpoint),
		&response,
		nil,
		nil,
	)
	if err != nil {
		return err
	}

	c.currentUser = response.Username

	return nil
}

// ListOrganizations return organizations for the current user.
func (c *Client) ListOrganizations(ctx context.Context, pVars *PaginationVars) ([]Organization, string, error) {
	var response ListResponse[Organization]

	err := c.doRequest(
		ctx,
		http.MethodGet,
		c.composeURL(UserOrgsEndpoint, c.currentUser),
		&response,
		nil,
		nil,
	)
	if err != nil {
		return nil, "", err
	}

	return response.Results, parsePageFromURL(response.Next), nil
}

// ListUsers return users under the provided organization.
func (c *Client) ListUsers(ctx context.Context, orgSlug string, pVars *PaginationVars) ([]User, string, error) {
	var response ListResponse[User]

	err := c.doRequest(
		ctx,
		http.MethodGet,
		c.composeURL(UsersEndpoint, orgSlug),
		&response,
		nil,
		pVars,
	)
	if err != nil {
		return nil, "", err
	}

	return response.Results, parsePageFromURL(response.Next), nil
}

// ListTeams return teams under the provided organization.
func (c *Client) ListTeams(ctx context.Context, orgSlug string, pVars *PaginationVars) ([]Team, string, error) {
	var response ListResponse[Team]

	err := c.doRequest(
		ctx,
		http.MethodGet,
		c.composeURL(TeamsEndpoint, orgSlug),
		&response,
		nil,
		pVars,
	)
	if err != nil {
		return nil, "", err
	}

	return response.Results, parsePageFromURL(response.Next), nil
}

// GetTeam return team details.
func (c *Client) GetTeam(ctx context.Context, orgSlug, teamId string) (*Team, error) {
	var response Team

	err := c.doRequest(
		ctx,
		http.MethodGet,
		c.composeURL(TeamDetailEndpoint, orgSlug, teamId),
		&response,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// ListTeamMembers return team members.
func (c *Client) ListTeamMembers(ctx context.Context, orgSlug, teamSlug string, pVars *PaginationVars) ([]User, string, error) {
	var response ListResponse[User]

	err := c.doRequest(
		ctx,
		http.MethodGet,
		c.composeURL(TeamMembersEndpoint, orgSlug, teamSlug),
		&response,
		nil,
		pVars,
	)
	if err != nil {
		return nil, "", err
	}

	return response.Results, parsePageFromURL(response.Next), nil
}

// ListRepositories return repositories under the provided organization.
func (c *Client) ListRepositories(ctx context.Context, orgSlug string, pVars *PaginationVars) ([]Repository, string, error) {
	var response ListResponse[Repository]

	err := c.doRequest(
		ctx,
		http.MethodGet,
		c.composeURL(RepositoriesEndpoint, orgSlug),
		&response,
		nil,
		pVars,
	)
	if err != nil {
		return nil, "", err
	}

	return response.Results, parsePageFromURL(response.Next), nil
}

// ListTeamPermissions return team permissions on provided repository.
func (c *Client) ListRepositoryPermissions(ctx context.Context, orgSlug, repoSlug string, pVars *PaginationVars) ([]RepositoryPermission, string, error) {
	var response ListResponse[RepositoryPermission]

	err := c.doRequest(
		ctx,
		http.MethodGet,
		c.composeURL(RepositoryPermissions, orgSlug, repoSlug),
		&response,
		nil,
		pVars,
	)
	if err != nil {
		return nil, "", err
	}

	return response.Results, parsePageFromURL(response.Next), nil
}

func setupPagination(query *url.Values, paginationVars *PaginationVars) {
	if paginationVars == nil {
		return
	}

	// add page size
	if paginationVars.Size != 0 {
		query.Set("page_size", fmt.Sprintf("%d", paginationVars.Size))
	}

	// add page
	if paginationVars.Page != "" {
		query.Set("page", paginationVars.Page)
	}
}

func (c *Client) doRequest(
	ctx context.Context,
	method string,
	urlAddress *url.URL,
	response interface{},
	data interface{},
	paginationVars *PaginationVars,
) error {
	var body []byte
	var err error

	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		urlAddress.String(),
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	q := url.Values{}
	setupPagination(&q, paginationVars)
	if q != nil {
		req.URL.RawQuery = q.Encode()
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	rawResponse, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer rawResponse.Body.Close()

	if rawResponse.StatusCode >= 300 {
		//nolint:gosec // No risk of overflow because `Code` is a small enum.
		return status.Error(codes.Code(rawResponse.StatusCode), "Request failed")
	}

	if err := json.NewDecoder(rawResponse.Body).Decode(&response); err != nil {
		return err
	}

	return nil
}

func parsePageFromURL(urlPayload string) string {
	if urlPayload == "" {
		return ""
	}

	u, err := url.Parse(urlPayload)
	if err != nil {
		return ""
	}

	return u.Query().Get("page")
}

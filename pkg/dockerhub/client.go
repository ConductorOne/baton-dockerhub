package dockerhub

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const (
	BaseDomain = "hub.docker.com"

	LoginEndpoint = "/v2/users/login"

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
	httpClient  *uhttp.BaseHttpClient
	baseUrl     *url.URL
	currentUser string

	username     string
	accessToken  string
	password     string
	token        string
	refreshToken string
}

type CredentialsReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenResp struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func NewClient(ctx context.Context, username, password, accessToken string) (*Client, error) {
	l := ctxzap.Extract(ctx)

	base := &url.URL{
		Scheme: "https",
		Host:   BaseDomain,
	}

	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, ctxzap.Extract(ctx)))
	if err != nil {
		return nil, err
	}

	wrapper, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, err
	}

	// Use username + password/accesstoken to get an auth token and refresh token
	credentials := CredentialsReq{
		Username: username,
		Password: password,
	}
	if password == "" {
		credentials.Password = accessToken
	}

	reqOptions := []uhttp.RequestOption{
		uhttp.WithContentType("application/json"),
		uhttp.WithAccept("application/json"),
		uhttp.WithJSONBody(credentials),
	}

	urlAddress := base.ResolveReference(&url.URL{Path: LoginEndpoint})

	req, err := wrapper.NewRequest(ctx, http.MethodGet, urlAddress, reqOptions...)
	if err != nil {
		return nil, err
	}

	data := &TokenResp{}
	doOptions := []uhttp.DoOption{
		uhttp.WithJSONResponse(data),
	}
	resp, err := wrapper.Do(req, doOptions...)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	l.Info("got token", zap.String("token", data.Token))

	return &Client{
		httpClient:   wrapper,
		baseUrl:      base,
		username:     username,
		accessToken:  accessToken,
		password:     password,
		token:        data.Token,
		refreshToken: data.RefreshToken,
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
func (c *Client) SetCurrentUser(ctx context.Context, username string) error {
	c.currentUser = username

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

func setupPagination(ctx context.Context, addr *url.URL, paginationVars *PaginationVars) *url.Values {
	if paginationVars == nil {
		return nil
	}

	q := addr.Query()

	// add page size
	if paginationVars.Size != 0 {
		q.Set("page_size", fmt.Sprintf("%d", paginationVars.Size))
	}

	// add page
	if paginationVars.Page != "" {
		q.Set("page", paginationVars.Page)
	}

	return &q
}

func (c *Client) doRequest(
	ctx context.Context,
	method string,
	urlAddress *url.URL,
	response interface{},
	data interface{},
	paginationVars *PaginationVars,
) error {
	var err error

	reqOptions := []uhttp.RequestOption{
		uhttp.WithContentType("application/json"),
		uhttp.WithAccept("application/json"),
		uhttp.WithBearerToken(c.token),
	}

	if data != nil {
		reqOptions = append(reqOptions, uhttp.WithJSONBody(data))
	}

	q := setupPagination(ctx, urlAddress, paginationVars)
	if q != nil {
		urlAddress.RawQuery = q.Encode()
	}

	req, err := c.httpClient.NewRequest(ctx, method, urlAddress, reqOptions...)
	if err != nil {
		return err
	}

	doOptions := []uhttp.DoOption{}
	if response != nil {
		doOptions = append(doOptions, uhttp.WithJSONResponse(response))
	}

	resp, err := c.httpClient.Do(req, doOptions...)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

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

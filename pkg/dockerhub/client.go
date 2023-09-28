package dockerhub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/conductorone/baton-dockerhub/pkg/dockerhub/external"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	BaseDomain = "hub.docker.com"

	OrgsEndpoint      = "/v2/orgs"
	OrgDetailEndpoint = OrgsEndpoint + "/%s"
	UsersEndpoint     = OrgsEndpoint + "/%s/members"

	TeamsEndpoint       = OrgsEndpoint + "/%s/groups"
	TeamDetailEndpoint  = TeamsEndpoint + "/%s"
	TeamMembersEndpoint = TeamDetailEndpoint + "/members"
)

type Client struct {
	httpClient *http.Client
	baseUrl    *url.URL

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
	Size int
	Page int
}

type ListResponse[T any] struct {
	PaginationData
	Results []T `json:"results"`
}

func (c *Client) GetUsers(ctx context.Context, organization string, pVars *PaginationVars) ([]User, string, error) {
	var response ListResponse[User]

	err := c.doRequest(
		ctx,
		http.MethodGet,
		c.composeURL(UsersEndpoint, organization),
		&response,
		nil,
		pVars,
	)
	if err != nil {
		return nil, "", err
	}

	return response.Results, response.Next, nil
}

func (c *Client) GetTeams(ctx context.Context, organization string) ([]Team, error) {
	var response ListResponse[Team]

	err := c.doRequest(
		ctx,
		http.MethodGet,
		c.composeURL(TeamsEndpoint, organization),
		&response,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return response.Results, nil
}

func (c *Client) GetTeamMembers(ctx context.Context, organization, team string) ([]TeamMember, error) {
	var response ListResponse[TeamMember]

	err := c.doRequest(
		ctx,
		http.MethodGet,
		c.composeURL(TeamMembersEndpoint, organization, team),
		&response,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return response.Results, nil
}

func setupPagination(query *url.Values, paginationVars *PaginationVars) {
	if paginationVars == nil {
		return
	}

	// add page size
	if paginationVars.Size != 0 {
		query.Set("page_size", strconv.Itoa(paginationVars.Size))
	}

	// add page
	if paginationVars.Page != 0 {
		query.Set("page", strconv.Itoa(paginationVars.Page))
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

	if data != nil {
		var err error
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
		return status.Error(codes.Code(rawResponse.StatusCode), "Request failed")
	}

	if err := json.NewDecoder(rawResponse.Body).Decode(&response); err != nil {
		return err
	}

	return nil
}

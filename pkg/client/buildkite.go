package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/oauth2"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

const (
	APIDomain      = "api.buildkite.com"
	GraphQLDomain  = "graphql.buildkite.com"
	APIVersion     = "v2"
	GraphQLVersion = "v1"
	MaxPageSize    = "100"
)

type Query struct {
	Query     string            `json:"query"`
	Variables map[string]string `json:"variables,omitempty"`
}

type Organization struct {
	Id           string    `json:"id"`
	GraphqlId    string    `json:"graphql_id"`
	Url          string    `json:"url"`
	WebUrl       string    `json:"web_url"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	PipelinesUrl string    `json:"pipelines_url"`
	AgentsUrl    string    `json:"agents_url"`
	EmojisUrl    string    `json:"emojis_url"`
	CreatedAt    time.Time `json:"created_at"`
}

type InfoQueryResponse struct {
	Data struct {
		Viewer struct {
			User *User `json:"user"`
		} `json:"viewer"`
	} `json:"data"`
}

type UsersQueryResponse struct {
	Data struct {
		Organization struct {
			Members struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Role string `json:"role"`
						User *User  `json:"user"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"members"`
		} `json:"organization"`
	} `json:"data"`
}

type TeamsQueryResponse struct {
	Data struct {
		Organization struct {
			Teams struct {
				Edges []struct {
					Node *Team `json:"node"`
				} `json:"edges"`
			} `json:"teams"`
		} `json:"organization"`
	} `json:"data"`
}

type TeamMembersQueryResponse struct {
	Data struct {
		Team struct {
			Members struct {
				Edges []struct {
					Node struct {
						Role string `json:"role"`
						ID   string `json:"id"`
						User *User  `json:"user"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"members"`
		} `json:"team"`
	} `json:"data"`
}

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// func (u *User) ConnectorUser() *connector_v1.User {
// 	cu := &connector_v1.User{
// 		Id:          u.ID,
// 		DisplayName: u.Name,
// 		Email:       u.Email,
// 		Profile:     nil,
// 		Status: &connector_v1.UserStatus{
// 			Status: connector_v1.UserStatus_STATUS_ENABLED,
// 		},
// 		Annotations: nil,
// 	}
// 	return cu
// }

type Team struct {
	OrgSlug string
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	ID      string `json:"id"`
}

// func (t *Team) Resource(res *connector_v1.ResourceType) *connector_v1.Resource {
// 	return &connector_v1.Resource{
// 		ResourceType: res,
// 		Id:           fmt.Sprintf("%s/%s", t.OrgSlug, t.Slug),
// 		DisplayName:  t.Name,
// 	}
// }

type TeamGrant struct {
	TeamID      string
	PrincipalID string
}

// func (r *TeamGrant) Grant(entitlement *connector_v1.Entitlement, resource *connector_v1.Resource) *connector_v1.Grant {
// 	principal := &connector_v1.Principal{
// 		Id:     r.PrincipalID,
// 		TypeId: connector.PrincipalTypeUser,
// 	}
// 	return &connector_v1.Grant{
// 		Entitlement: entitlement,
// 		Id:          fmt.Sprintf("grant:team:%s:%s", r.TeamID, r.PrincipalID),
// 		Principal:   principal,
// 	}
// }

// type Resource interface {
// 	Resource(res *connector_v1.ResourceType) *connector_v1.Resource
// }

// type Grant interface {
// 	Grant(entitlement *connector_v1.Entitlement, resource *connector_v1.Resource) *connector_v1.Grant
// }

type InfoResponse struct {
	User                 *User
	RateLimitDescription *v2.RateLimitDescription
}

type UsersResponse struct {
	Users                []*User
	RateLimitDescription *v2.RateLimitDescription
	Pagination           string
}

type TeamsGrantsResponse struct {
	Grants               []*TeamGrant
	RateLimitDescription *v2.RateLimitDescription
	Pagination           string
}

type TeamsResponse struct {
	Teams                []*Team
	RateLimitDescription *v2.RateLimitDescription
	Pagination           string
}

type Client interface {
	GetInfo(ctx context.Context) (*InfoResponse, error)
	ListUsers(ctx context.Context, orgSlug string, pagination string) (*UsersResponse, error)
	ListTeams(ctx context.Context, orgSlug string, pagination string) (*TeamsResponse, error)
	ListTeamGrants(ctx context.Context, teamOrgSlug string, pagination string) (*TeamsGrantsResponse, error)
	ListOrganizations(ctx context.Context) ([]*Organization, error)
}

type connectorClient struct {
	client      *uhttp.BaseHttpClient
	tokenSource oauth2.TokenSource
}

const (
	userAgent = "ConductorOne/buildkite-connector-0.2.0"
)

func New(ctx context.Context, ts oauth2.TokenSource) (Client, error) {
	httpClient, err := uhttp.NewClient(
		ctx,
		uhttp.WithLogger(true, ctxzap.Extract(ctx)),
		uhttp.WithUserAgent(userAgent),
	)
	if err != nil {
		return nil, err
	}
	wrapper, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, err
	}

	return &connectorClient{
		client:      wrapper,
		tokenSource: ts,
	}, nil
}

func (c *connectorClient) newUnPaginatedURL(path string, v url.Values) (string, error) {
	reqUrl, err := url.Parse(fmt.Sprintf("https://%s/%s/%s", APIDomain, APIVersion, path))
	if err != nil {
		return "", err
	}
	if v == nil {
		v = url.Values{}
	}
	v.Set("page", "1")
	v.Set("per_page", MaxPageSize)
	reqUrl.RawQuery = v.Encode()
	return reqUrl.String(), nil
}

func (c *connectorClient) req(ctx context.Context, method string, requestURL string, res interface{}) error {
	reqUrl, err := url.Parse(requestURL)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, reqUrl.String(), nil)
	if err != nil {
		return err
	}
	token, err := c.tokenSource.Token()
	if err != nil {
		return err
	}
	req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", token.AccessToken)}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("buildkite-client: Rest API HTTP request failed %d %s", resp.StatusCode, string(rawResp))
	}
	if err := json.Unmarshal(rawResp, res); err != nil {
		return err
	}
	return nil
}

func (c *connectorClient) query(ctx context.Context, rawQuery string, res interface{}) (*v2.RateLimitDescription, error) {
	reqUrl := &url.URL{
		Scheme: "https",
		Host:   GraphQLDomain,
		Path:   path.Join("/", GraphQLVersion),
	}
	q := &Query{
		Query: rawQuery,
	}

	token, err := c.tokenSource.Token()
	if err != nil {
		return nil, err
	}

	spew.Dump(q)

	req, err := c.client.NewRequest(ctx, http.MethodPost, reqUrl,
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithJSONBody(q),
		uhttp.WithHeader("Authorization", "Bearer "+token.AccessToken),
	)
	if err != nil {
		return nil, err
	}

	rld := &v2.RateLimitDescription{}
	httpResp, err := c.client.Do(req,
		uhttp.WithJSONResponse(res),
		uhttp.WithRatelimitData(rld),
	)
	if err != nil {
		return nil, err
	}
	spew.Dump(httpResp.StatusCode)
	return rld, nil
}

func (c *connectorClient) ListOrganizations(ctx context.Context) ([]*Organization, error) {
	orgs := make([]*Organization, 0)
	reqURL, err := c.newUnPaginatedURL("organizations", nil)
	if err != nil {
		return nil, err
	}
	if err := c.req(ctx, http.MethodGet, reqURL, &orgs); err != nil {
		return nil, err
	}
	return orgs, nil
}

func (c *connectorClient) GetInfo(ctx context.Context) (*InfoResponse, error) {
	resp := &InfoQueryResponse{}
	rld, err := c.query(ctx, infoQuery(), resp)
	if err != nil {
		return nil, fmt.Errorf("buildkite-client: error getting info %w", err)
	}

	rv := &InfoResponse{
		User:                 resp.Data.Viewer.User,
		RateLimitDescription: rld,
	}
	return rv, nil
}

func (c *connectorClient) ListUsers(ctx context.Context, orgSlug string, pagination string) (*UsersResponse, error) {
	resp := &UsersQueryResponse{}
	rld, err := c.query(ctx, allUsersQuery(orgSlug, pagination), resp)
	if err != nil {
		return nil, fmt.Errorf("buildkite-client: error getting all users %w", err)
	}
	spew.Dump(resp)
	users := make([]*User, 0, len(resp.Data.Organization.Members.Edges))
	lastIndex := -1
	for i, edge := range resp.Data.Organization.Members.Edges {
		users = append(users, edge.Node.User)
		lastIndex = i
	}
	pg := ""
	if lastIndex > -1 {
		pg = resp.Data.Organization.Members.Edges[lastIndex].Node.ID
	}
	rv := &UsersResponse{
		Users:                users,
		RateLimitDescription: rld,
		Pagination:           pg,
	}
	return rv, nil
}

func (c *connectorClient) ListTeams(ctx context.Context, orgSlug string, pagination string) (*TeamsResponse, error) {
	resp := &TeamsQueryResponse{}
	rld, err := c.query(ctx, teamsQuery(orgSlug, pagination), resp)
	if err != nil {
		return nil, fmt.Errorf("buildkite-client: error getting teams %w", err)
	}
	teams := make([]*Team, 0, len(resp.Data.Organization.Teams.Edges))
	lastIndex := -1
	for i, edge := range resp.Data.Organization.Teams.Edges {
		team := edge.Node
		team.OrgSlug = orgSlug
		teams = append(teams, team)
		lastIndex = i
	}
	pg := ""
	if lastIndex > -1 {
		pg = resp.Data.Organization.Teams.Edges[lastIndex].Node.ID
	}
	rv := &TeamsResponse{
		Teams:                teams,
		RateLimitDescription: rld,
		Pagination:           pg,
	}
	return rv, nil
}

func (c *connectorClient) ListTeamGrants(ctx context.Context, teamOrgSlug string, pagination string) (*TeamsGrantsResponse, error) {
	resp := &TeamMembersQueryResponse{}
	rld, err := c.query(ctx, teamMembersQuery(teamOrgSlug, pagination), resp)
	if err != nil {
		return nil, fmt.Errorf("buildkite-client: error getting team members for %s: %w", teamOrgSlug, err)
	}
	grants := make([]*TeamGrant, 0, len(resp.Data.Team.Members.Edges))
	lastIndex := -1
	for i, edge := range resp.Data.Team.Members.Edges {
		grants = append(grants, &TeamGrant{
			TeamID:      teamOrgSlug,
			PrincipalID: edge.Node.User.ID,
		})
		lastIndex = i
	}
	pg := ""
	if lastIndex > -1 {
		pg = resp.Data.Team.Members.Edges[lastIndex].Node.ID
	}
	rv := &TeamsGrantsResponse{
		Grants:               grants,
		RateLimitDescription: rld,
		Pagination:           pg,
	}
	return rv, nil
}

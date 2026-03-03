package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

const (
	defaultBaseURL = "https://api.buildkite.com/v2"
	perPage        = "100"
)

var linkNextRe = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
	org        string
}

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Member struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Team struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Description   string `json:"description"`
	Privacy       string `json:"privacy"`
	IsDefaultTeam bool   `json:"default"`
	CreatedAt     string `json:"created_at"`
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Role     string `json:"role"`
}

func New(ctx context.Context, token, org, baseURL string) (*Client, error) {
	httpClient, err := uhttp.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("baton-buildkite: failed to create http client: %w", err)
	}

	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		token:      token,
		org:        org,
	}, nil
}

func (c *Client) ListOrganizations(ctx context.Context) ([]Organization, error) {
	url := fmt.Sprintf("%s/organizations", c.baseURL)

	var orgs []Organization
	_, err := c.doRequest(ctx, url, &orgs)
	if err != nil {
		return nil, fmt.Errorf("baton-buildkite: failed to list organizations: %w", err)
	}

	return orgs, nil
}

func (c *Client) ListMembers(ctx context.Context, pageURL string) ([]Member, string, error) {
	if pageURL == "" {
		pageURL = fmt.Sprintf("%s/organizations/%s/members?per_page=%s", c.baseURL, c.org, perPage)
	}

	var members []Member
	nextPage, err := c.doRequest(ctx, pageURL, &members)
	if err != nil {
		return nil, "", fmt.Errorf("baton-buildkite: failed to list members: %w", err)
	}

	return members, nextPage, nil
}

func (c *Client) ListTeams(ctx context.Context, pageURL string) ([]Team, string, error) {
	if pageURL == "" {
		pageURL = fmt.Sprintf("%s/organizations/%s/teams?per_page=%s", c.baseURL, c.org, perPage)
	}

	var teams []Team
	nextPage, err := c.doRequest(ctx, pageURL, &teams)
	if err != nil {
		return nil, "", fmt.Errorf("baton-buildkite: failed to list teams: %w", err)
	}

	return teams, nextPage, nil
}

func (c *Client) ListTeamMembers(ctx context.Context, teamID, pageURL string) ([]TeamMember, string, error) {
	if pageURL == "" {
		pageURL = fmt.Sprintf("%s/organizations/%s/teams/%s/members?per_page=%s", c.baseURL, c.org, teamID, perPage)
	}

	var members []TeamMember
	nextPage, err := c.doRequest(ctx, pageURL, &members)
	if err != nil {
		return nil, "", fmt.Errorf("baton-buildkite: failed to list team members: %w", err)
	}

	return members, nextPage, nil
}

func (c *Client) doRequest(ctx context.Context, url string, target interface{}) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("baton-buildkite: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req) //nolint:gosec // URL is constructed from trusted config, not user input
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return "", fmt.Errorf("baton-buildkite: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("baton-buildkite: unexpected status %d for %s", resp.StatusCode, url)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return "", fmt.Errorf("baton-buildkite: failed to decode response: %w", err)
	}

	return parseLinkNext(resp.Header.Get("Link")), nil
}

func parseLinkNext(header string) string {
	if header == "" {
		return ""
	}

	matches := linkNextRe.FindStringSubmatch(header)
	if len(matches) < 2 {
		return ""
	}

	return matches[1]
}

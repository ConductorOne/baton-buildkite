package client

import "fmt"

const (
	perPageQuery = 100

	firstPagePagination = "first: %d"
	nextPagePagination  = "first: %d after: %s"

	getAllUsersQuery = `query getOrgMembers{
  organization(slug: "%s") {
    members(%s ) {
      edges {
        node {
          role
          user {
            name
            email
            id
          }
        }
      }
    }
  }
}`

	getInfoQuery = `query getInfo {
  viewer {
    user{
      id
      email
      name
    }
  }
}
`

	getTeamsQuery = `query getTeams {
  organization(slug: "%s") {
    teams(%s) {
      edges {
        node {
          slug
          name
          id
        }
      }
    }
  }
}`

	getTeamMemberListQuery = `query getTeamMembers {
  team(slug: "%s") {
    members(%s) {
      edges {
        node {
          role
          id
          user {
            name
            email
            id
          }
        }
      }
    }
  }
}`
)

func pagination(pg string) string {
	if pg == "" {
		return fmt.Sprintf(firstPagePagination, perPageQuery)
	}
	return fmt.Sprintf(nextPagePagination, perPageQuery, pg)
}

func allUsersQuery(orgSlug string, pg string) string {
	return fmt.Sprintf(getAllUsersQuery, orgSlug, pagination(pg))
}

func teamsQuery(orgSlug string, pg string) string {
	return fmt.Sprintf(getTeamsQuery, orgSlug, pagination(pg))
}

func infoQuery() string {
	return getInfoQuery
}

func teamMembersQuery(teamOrgSlug string, pg string) string {
	return fmt.Sprintf(getTeamMemberListQuery, teamOrgSlug, pagination(pg))
}

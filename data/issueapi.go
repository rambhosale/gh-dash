package data

import (
	"fmt"
	"log"
	"time"

	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
)

type IssueData struct {
	Number int
	Title  string
	Body   string
	State  string
	Author struct {
		Login string
	}
	UpdatedAt  time.Time
	Url        string
	Repository struct {
		Name          string
		NameWithOwner string
	}
	Assignees Assignees      `graphql:"assignees(first: 3)"`
	Comments  Comments       `graphql:"comments(first: 15)"`
	Reactions IssueReactions `graphql:"reactions(first: 1)"`
	Labels    IssueLabels    `graphql:"labels(first: 3)"`
}

type PageInfo struct {
	HasNextPage bool   `graphql:"hasNextPage"`
	EndCursor   string `graphql:"endCursor"`
}

type Assignees struct {
	Nodes []Assignee
}

type Assignee struct {
	Login string
}

type IssueReactions struct {
	TotalCount int
}

type Label struct {
	Color string
	Name  string
}

type IssueLabels struct {
	Nodes []Label
}

func (data IssueData) GetRepoNameWithOwner() string {
	return data.Repository.NameWithOwner
}

func (data IssueData) GetNumber() int {
	return data.Number
}

func (data IssueData) GetUrl() string {
	return data.Url
}

func (data IssueData) GetUpdatedAt() time.Time {
	return data.UpdatedAt
}

func makeIssuesQuery(query string) string {
	return fmt.Sprintf("is:issue %s", query)
}

func FetchIssues(query string, limit int, cursor PageInfo) ([]IssueData, PageInfo, error) {
	var err error
	client, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
		return nil, PageInfo{}, err
	}

	var queryResult struct {
		Search struct {
			Nodes []struct {
				Issue IssueData `graphql:"... on Issue"`
			}
			Cursor PageInfo `graphql:"pageInfo"`
		} `graphql:"search(type: ISSUE, first: $limit, after: $endCursor, query: $query)"`
	}

	variables := map[string]interface{}{
		"query":     graphql.String(makeIssuesQuery(query)),
		"limit":     graphql.Int(limit),
		"endCursor": (*graphql.String)(nil),
	}

	if cursor.HasNextPage {
		variables["endCursor"] = graphql.String(cursor.EndCursor)
	}

	err = client.Query("SearchIssues", &queryResult, variables)
	if err != nil {
		return nil, PageInfo{}, err
	}

	issues := make([]IssueData, 0, len(queryResult.Search.Nodes))
	for _, node := range queryResult.Search.Nodes {
		issues = append(issues, node.Issue)
	}

	return issues, queryResult.Search.Cursor, nil
}

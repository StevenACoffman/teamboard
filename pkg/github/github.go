package github

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/StevenACoffman/teamboard/pkg/generated/genqlient"
	"github.com/StevenACoffman/teamboard/pkg/types"
	"sort"
	"strings"
)

func GetLogin(
	ctx context.Context,
	graphqlClient graphql.Client,
) (string, error) {
	myLoginResp, err := genqlient.MyLogin(ctx, graphqlClient)
	if err != nil  {
		return "", err
	}
	return  myLoginResp.Viewer.Login, nil
}
func GetTeams(ctx context.Context, graphqlClient graphql.Client, myLogin, org string) ([]string, error){
	var myTeams []string

	myTeamsResp, err := genqlient.MyTeams(ctx, graphqlClient, org, myLogin)

	if err != nil {
		return nil, err
	}
	for _, edge := range myTeamsResp.Organization.Teams.Edges {
		myTeams = append(myTeams, edge.Node.Name)
	}
	return myTeams, nil
}
func GetOrgs(ctx context.Context, graphqlClient graphql.Client) ([]string, error){
	var myOrgs []string

	myOrgsResp, err := genqlient.MyOrgs(ctx, graphqlClient)
	if err != nil {
		return nil, err
	}
	for _, node := range myOrgsResp.Viewer.Organizations.Nodes {
		myOrgs = append(myOrgs, node.Login)
	}
	return myOrgs, nil
}
func GetTeamMembers(ctx context.Context, graphqlClient graphql.Client, org string, team string) ( []string, error) {
	fmt.Println("Getting team members for org: ", org, " team:", team)

	var teammates []string
	teamResp, teamErr := genqlient.TeamMembers(ctx, graphqlClient, org, team)
	if teamErr != nil {
		return nil, teamErr
	}

	for _, team := range teamResp.Organization.Teams.Edges {
		for _, edge := range team.Node.Members.Edges {
			teammates = append(teammates, edge.Node.Login)
		}
	}
	return teammates, nil
}

func GetPulls(
	ctx context.Context,
	graphqlClient graphql.Client,
	myLogin string,
	org string,
	team string,
	teammates []string,
) ([]types.PullRequest, error) {

	teamAuthoredQuery := fmt.Sprintf(
		"is:open is:pr org:%s archived:false draft:false author:%s",
		org,
		strings.Join(teammates, " author:"),
	)

	meRequestedQuery := fmt.Sprintf(
		"is:open is:pr org:%s  archived:false review-requested:%s",
		org,
		myLogin,
	)
	meMentionedQuery := fmt.Sprintf(
		"is:open is:pr org:%s archived:false mentions:%s",
		org,
		myLogin,
	)

	teamMentionedQuery := fmt.Sprintf(
		"is:open is:pr org:%s archived:false team:%s/%s",
		org,
		org,
		team,
	)
	teamRequestedQuery := fmt.Sprintf(
		"is:open is:pr org:%s archived:false team-review-requested:%s/%s",
		org,
		org,
		team,
	)

	resp, err := genqlient.MyBatch(
		ctx,
		graphqlClient,
		meRequestedQuery,
		meMentionedQuery,
		teamAuthoredQuery,
		teamMentionedQuery,
		teamRequestedQuery,
	)
	var pulls []types.PullRequest

	if err != nil {
		return nil, err
	}
	for _, edge := range resp.Mementioned.Edges {
		// s, ok :=
		// edge.Node.(*genqlient.MyBatchMementionedSearchResultItemConnectionEdgesSearchResultItemEdgeNodePullRequest)
		s, ok := edge.Node.(*types.PullRequest)
		if ok {
			pulls = append(pulls, *s)
		}
	}
	for _, edge := range resp.Merequested.Edges {

		s, ok := edge.Node.(*types.PullRequest)
		if ok {
			pulls = append(pulls, *s)
		}
	}
	for _, edge := range resp.Teammates.Edges {

		// s, ok :=
		// edge.Node.(*genqlient.MyBatchTeammatesSearchResultItemConnectionEdgesSearchResultItemEdgeNodePullRequest)
		s, ok := edge.Node.(*types.PullRequest)
		if ok {
			pulls = append(pulls, *s)
		}
	}
	for _, edge := range resp.Teammentions.Edges {

		// s, ok :=
		// edge.Node.(*genqlient.MyBatchTeammentionsSearchResultItemConnectionEdgesSearchResultItemEdgeNodePullRequest)
		s, ok := edge.Node.(*types.PullRequest)
		if ok {
			pulls = append(pulls, *s)
		}

	}
	for _, edge := range resp.Teamrequested.Edges {
		// s, ok := (&edge).(types.EdgeNodePullRequest)
		s, ok := edge.Node.(*types.PullRequest)
		if ok {
			pulls = append(pulls, *s)
		}
	}
	sort.Slice(pulls, func(i, j int) bool {
		// results in most recent to oldest
		return pulls[i].CreatedAt.After(pulls[j].CreatedAt)
	})
	// Do we really want to do this?
	// Or should we have added a reason / category to them above instead?
	pulls = removeDuplicateValues(pulls)
	return pulls, nil
}

func removeDuplicateValues(pulls []types.PullRequest) []types.PullRequest {
	keys := make(map[string]bool)
	var list []types.PullRequest

	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range pulls {
		if _, value := keys[entry.Url]; !value {
			keys[entry.Url] = true
			list = append(list, entry)
		}
	}
	return list
}

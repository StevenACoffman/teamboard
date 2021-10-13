package cmd

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/Khan/genqlient/graphql"
	"github.com/StevenACoffman/teamboard/pkg"
	"github.com/StevenACoffman/teamboard/pkg/middleware"
	"github.com/StevenACoffman/teamboard/pkg/types"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/StevenACoffman/teamboard/pkg/generated/genqlient"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "teamboard",
	Short: "Get Open GitHub Pull Requests for you and your Team",
	Long:  `This lets you `,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		defer func() {
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}()

		key := os.Getenv("GITHUB_TOKEN")
		if key == "" {
			err = fmt.Errorf("must set GITHUB_TOKEN=<github token>")
			return
		}

		httpClient := middleware.NewBearerAuthHTTPClient(key)
		ctx := context.Background()

		graphqlClient := graphql.NewClient("https://api.github.com/graphql", httpClient)
		org := "Khan"
		team := "districts"

		myLogin, myOrgs, myTeams, teammates, paramErr := GetParams(ctx, graphqlClient, org, team)
		if paramErr != nil {
			err = paramErr
			return
		}

		fmt.Println("<!-- myTeams: ", myTeams, " -->")
		fmt.Println("<!-- myOrgs: ", myOrgs, " -->")
		fmt.Println("<!-- TeamMates: ", teammates, " -->")
		pulls, pullsErr := GetPulls(ctx, graphqlClient, myLogin, org, team, teammates)
		if pullsErr != nil {
			fmt.Printf("%+v\n", pullsErr)
			return
		}

		t, tmplParseErr := template.ParseFS(pkg.AssetData, "assets/team-pr-template.html")
		if tmplParseErr != nil {
			fmt.Println(tmplParseErr)
		}

		buf := &bytes.Buffer{}
		if err := t.Execute(buf, pulls); err != nil {
			panic(err)
		}
		fragment := buf.String()
		fmt.Println(fragment)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().
		StringVar(&cfgFile, "config", "", "config file (default is $HOME/.teamboard.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".teamboard" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".teamboard")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func GetParams(
	ctx context.Context,
	graphqlClient graphql.Client,
	org string,
	team string,
) (myLogin string,
	myOrgs []string,
	myTeams []string,
	teammates []string,
	err error) {

	myLoginResp, myLoginErr := genqlient.MyLogin(ctx, graphqlClient)
	if myLoginErr != nil {
		err = myLoginErr
		return
	}
	myLogin = myLoginResp.Viewer.Login

	myTeamsResp, myTeamsErr := genqlient.MyTeams(ctx, graphqlClient, org, myLogin)

	if myTeamsErr != nil {
		err = myTeamsErr
		return
	}
	for _, edge := range myTeamsResp.Organization.Teams.Edges {
		myTeams = append(myTeams, edge.Node.Name)
	}

	myOrgsResp, myOrgsErr := genqlient.MyOrgs(ctx, graphqlClient)
	if myOrgsErr != nil {
		err = myOrgsErr
		return
	}
	for _, node := range myOrgsResp.Viewer.Organizations.Nodes {
		myOrgs = append(myOrgs, node.Login)
	}

	teamResp, teamErr := genqlient.TeamMembers(ctx, graphqlClient, org, team)
	if teamErr != nil {
		err = teamErr
		return
	}

	for _, team := range teamResp.Organization.Teams.Edges {
		for _, edge := range team.Node.Members.Edges {
			teammates = append(teammates, edge.Node.Login)
		}
	}

	return
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
		"is:open is:pr is:private archived:false draft:false author:%s",
		strings.Join(teammates, " author:"),
	)

	meRequestedQuery := fmt.Sprintf(
		"is:open is:pr is:private archived:false review-requested:%s",
		myLogin,
	)
	meMentionedQuery := fmt.Sprintf(
		"is:open is:pr is:private archived:false mentions:%s",
		myLogin,
	)

	teamMentionedQuery := fmt.Sprintf(
		"is:open is:pr is:private archived:false team:%s/%s",
		org,
		team,
	)
	teamRequestedQuery := fmt.Sprintf(
		"is:open is:pr is:private archived:false team-review-requested:%s/%s",
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
	return pulls, nil
}

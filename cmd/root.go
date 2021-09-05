package cmd

import (
  "context"
  "fmt"
  "github.com/Khan/genqlient/graphql"
    "github.com/StevenACoffman/teamboard/pkg/middleware"
    "github.com/spf13/cobra"
  "net/http"
  "os"

  homedir "github.com/mitchellh/go-homedir"
  "github.com/spf13/viper"

  "github.com/StevenACoffman/teamboard/pkg/generated/genqlient"
)

var cfgFile string

type authedTransport struct {
  key     string
  wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
  req.Header.Set("Authorization", "bearer "+t.key)
  return t.wrapped.RoundTrip(req)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
  Use:   "teamboard",
  Short: "Get Open GitHub Pull Requests for you and your Team",
  Long: `This lets you `,
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

      graphqlClient := graphql.NewClient("https://api.github.com/graphql", httpClient)
      meRequestedQuery:= "is:open is:pr is:private archived:false review-requested:StevenACoffman"
      meMentionedQuery:= "is:open is:pr is:private archived:false mentions:StevenACoffman"
      teamAuthoredQuery:= "is:open is:pr is:private archived:false author:StevenACoffman author:drewkiimon author:kphilipkhan author:jeffkhan"
      teamMentionedQuery:= "is:open is:pr is:private archived:false team:Khan/districts"
      teamRequestedQuery:= "is:open is:pr is:private archived:false team-review-requested:Khan/districts"
      var resp *genqlient.MyBatchResponse

      resp, err = genqlient.MyBatch(context.Background(), graphqlClient, meRequestedQuery, meMentionedQuery, teamAuthoredQuery, teamMentionedQuery, teamRequestedQuery)


        if err != nil {
          return
        }
        for _, edge := range resp.Mementioned.Edges {
 // edge is MyBatchMementionedSearchResultItemConnectionEdgesSearchResultItemEdge
 // not sure how to get to MyBatchMerequestedSearchResultItemConnectionEdgesSearchResultItemEdgeNodePullRequest

            fmt.Println(edge.Node.GetTypename())
        }

        fmt.Printf("%+v\n", resp.Mementioned.Edges)
        fmt.Printf("%+v\n",resp.Merequested.Edges)
        fmt.Printf("%+v\n",resp.Teammates.Edges)
        fmt.Printf("%+v\n",resp.Teammentions.Edges)
        fmt.Printf("%+v\n",resp.Teamrequested.Edges)
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

  rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.teamboard.yaml)")


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


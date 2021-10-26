package cmd

import (
	_ "embed"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/StevenACoffman/teamboard/pkg/middleware"
	"github.com/StevenACoffman/teamboard/pkg/server"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
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

		graphqlClient := graphql.NewClient("https://api.github.com/graphql", httpClient)

		// App Starting
		logger := log.New(os.Stdout,
			"INFO: ",
			log.Ldate|log.Ltime|log.Lshortfile)
		logger.Printf("main : Started")
		err = server.RunServer(logger, graphqlClient)
		if err == nil {
			logger.Println("finished clean")
			os.Exit(0)
		} else {
			logger.Printf("Got error: %v", err)
			os.Exit(1)
		}

		if err != nil {
			fmt.Println("ERROR", err)
		}
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


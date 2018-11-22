// Copyright © 2018 Developer developer@vamp.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/magneticio/vamp2cli/client"
	"github.com/spf13/cobra"
)

var Url string
var Username string
var Password string

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("Server: " + Server)
		// fmt.Println("login called for " + Username + " " + Password)
		// fmt.Println("Print: " + strings.Join(args, " "))
		if Url != "" {
			Config.Url = Url
		}
		if Config.Url == "" {
			fmt.Println("A Vamp Service url should be provided by url flag")
			return
		}
		restClient := client.NewRestClient(Config.Url, Config.Token)
		token, _ := restClient.Login(Username, Password)
		fmt.Println("Token will be written to config: " + token)
		Config.Token = token
		WriteConfigFile()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")
	// loginCmd.PersistentFlags().StringVar(&Server, "server", "", "Server to connect")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	loginCmd.Flags().StringVarP(&Url, "url", "", "", "Url required")
	loginCmd.Flags().StringVarP(&Username, "user", "u", "", "Username required")
	loginCmd.MarkFlagRequired("user")
	loginCmd.Flags().StringVarP(&Password, "password", "p", "", "Password required")
	loginCmd.MarkFlagRequired("password")

	// loginCmd.PersistentFlags().StringVar(&Server, "server", "default", "server to connect")
	// viper.BindPFlag("server", loginCmd.PersistentFlags().Lookup("server"))
	// Server = viper.GetString("server")
}

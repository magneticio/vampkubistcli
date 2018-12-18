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
	"errors"
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
	Short: "login to a vamp service",
	Long: `Login to a vamp service:
Example:
  vamp2cli --url https://1.2.3.4:8888 --user username --password password

  Login creates a configuration file in the home folder of the user.
  Username and password is not stored in the configuration, only token is stored.
  Default config location is ~/.vamp2cli.yaml
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// fmt.Println("Server: " + Server)
		// fmt.Println("login called for " + Username + " " + Password)
		// fmt.Println("Print: " + strings.Join(args, " "))
		if Url != "" {
			Config.Url = Url
		}
		if Config.Url == "" {
			// fmt.Println("A Vamp Service url should be provided by url flag")
			return errors.New("A Vamp Service url should be provided by url flag")
		}
		restClient := client.NewRestClient(Config.Url, Config.Token, Debug)
		token, err := restClient.Login(Username, Password)
		if err != nil {
			return err
		}
		fmt.Println("Token will be written to config: " + token)
		Config.Token = token
		WriteConfigFile()
		return nil
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
	loginCmd.Flags().StringVarP(&Username, "user", "", "", "Username required")
	loginCmd.MarkFlagRequired("user")
	loginCmd.Flags().StringVarP(&Password, "password", "", "", "Password required")
	loginCmd.MarkFlagRequired("password")

	// loginCmd.PersistentFlags().StringVar(&Server, "server", "default", "server to connect")
	// viper.BindPFlag("server", loginCmd.PersistentFlags().Lookup("server"))
	// Server = viper.GetString("server")
}

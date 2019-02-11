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
	"syscall"

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	//"golang.org/x/crypto/ssh/terminal"
)

var Url string
var Username string
var Password string
var Cert string

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to a vamp service",
	Long: AddAppName(`Login to a vamp service:
Example:
  Logging in with using username and password:
  $AppName login --url https://1.2.3.4:8888 --user username --password password
  Logging in with an existing token:
  $AppName login --url https://1.2.3.4:8888 --token dXNlcjE6ZnJvbnRlbmQ6MTU0NzU2MDc5ODcyMzo5OHJhcFRydEloZXBEVW1PV0F6UQ==

  It is also possible to pass certificate with cert parameter
  $AppName login --url https://1.2.3.4:8888 --user username --password password --cert file-or-string

  Interactive password input is enabled if username is entered
  but password is not passed for security:

  $AppName login --url https://1.2.3.4:8888 --user username

  Cert parameter accepts cerficate string, local file path or remote file path.

  Login creates a configuration file in the home folder of the user.
  Username and password is not stored in the configuration, only token is stored.
  Default config location is ~/.$AppName.yaml
`),
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
		CertString := Cert
		if Cert != "" {
			err_cert := util.VerifyCertForHost(Config.Url, Cert)
			if err_cert != nil {
				b, err := util.UseSourceUrl(Cert)
				if err != nil {
					fmt.Printf("Warning: %v\n", err)
					b = Cert
				}
				certVerifyError := util.VerifyCertForHost(Config.Url, b)
				if certVerifyError != nil {
					return certVerifyError
				}
				CertString = string(b)
			}
			Config.Cert = CertString
		}
		if Token != "" {
			Config.Token = Token
			restClient := client.NewRestClient(Config.Url, Config.Token, Debug, Config.Cert)
			isPong, err := restClient.Ping() // TODO: use an authorized endpoint to check token works
			if !isPong {
				return err
			}
		} else {
			if Username == "" {
				return errors.New("Username is required")
			}
			if Password == "" {
				fmt.Println("Enter your password (password will not be visible):")
				bytePassword, errInput := terminal.ReadPassword(int(syscall.Stdin))
				if errInput != nil {
					return errInput
				}
				Password = string(bytePassword)
				fmt.Println()
				if Password == "" {
					return errors.New("Password is required")
				}
			}
			restClient := client.NewRestClient(Config.Url, Config.Token, Debug, Config.Cert)
			token, err := restClient.Login(Username, Password)
			if err != nil {
				return err
			}
			Config.Token = token
		}
		fmt.Println("Token will be written to config: " + Config.Token)
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
	// loginCmd.MarkFlagRequired("user")
	loginCmd.Flags().StringVarP(&Password, "password", "", "", "Password required")
	// loginCmd.MarkFlagRequired("password")
	loginCmd.Flags().StringVarP(&Cert, "cert", "", "", "Cert from file, url or string")

	// loginCmd.PersistentFlags().StringVar(&Server, "server", "default", "server to connect")
	// viper.BindPFlag("server", loginCmd.PersistentFlags().Lookup("server"))
	// Server = viper.GetString("server")
}

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

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
)

// passwdCmd represents the passwd command
var passwdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "Update password",
	Long: AddAppName(`Update password
    For the current user:
    $AppName passwd
    For a different user
    $AppName passwd --user username`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		passwd, passwdError := util.GetParameterFromTerminalAsSecret(
			"Enter your password (password will not be visible):",
			"Enter your password again (password will not be visible):",
			"Passwords do not match.")
		if passwdError != nil {
			return passwdError
		}
		if Username == "" {
			Username = Config.Username
		}
		Source := "{\"userName\":\"" + Username + "\",\"password\":\"" + passwd + "\"}"
		restClient := client.NewRestClient(Config.Url, Config.Token, Config.APIVersion, logging.Verbose, Config.Cert, &TokenStore)
		values := make(map[string]string)
		values["project"] = Config.Project
		values["cluster"] = Config.Cluster
		values["virtual_cluster"] = Config.VirtualCluster
		values["application"] = Application
		updateError := restClient.UpdatePassword(Username, passwd, Source, values)
		if updateError != nil {
			return updateError
		}
		fmt.Printf("User password updated.\nLogin with the new password.\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(passwdCmd)

	passwdCmd.Flags().StringVarP(&Username, "user", "", "", "Username of the user to change password")
}

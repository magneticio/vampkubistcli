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

	"github.com/magneticio/forklift/logging"
	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
)

// passwdCmd represents the passwd command
var passwdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "Update password",
	Long: AddAppName(`Update password
    For the current user:
    $AppName
    For a different user
    $AppName --user username`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		Type := "user"
		passwd, passwdError := util.GetParameterFromTerminalAsSecret(
			"Enter your password (password will not be visible):",
			"Enter your password again (password will not be visible):",
			"Passwords do not match.")
		if passwdError != nil {
			return passwdError
		}
		SourceFileType := "json"
		if Username == "" {
			Username = Config.Username
		}
		Source := "{\"userName\":\"" + Username + "\",\"password\":\"" + passwd + "\"}"
		restClient := client.NewRestClient(Config.Url, Config.Token, Config.APIVersion, logging.Verbose, Config.Cert)
		values := make(map[string]string)
		values["project"] = Config.Project
		values["cluster"] = Config.Cluster
		values["virtual_cluster"] = Config.VirtualCluster
		values["application"] = Application
		_, updateError := restClient.Update(Type, Username, Source, SourceFileType, values)
		if updateError != nil {
			return updateError
		}
		fmt.Printf("User password updated.\nLogin with the new password.\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(passwdCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// passwdCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// passwdCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	passwdCmd.Flags().StringVarP(&Username, "user", "", "", "Username of the user to change password")
}

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
	"github.com/magneticio/vampkubistcli/client"
	"github.com/spf13/cobra"
)

var Role string

// grantCmd represents the grant command
var grantCmd = &cobra.Command{
	Use:   "grant",
	Short: "Grant a role or permission to a user for an object",
	Long: AddAppName(`Usage:
$AppName grant --user user1 --role admin -p default`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		restClient := client.NewRestClient(Config.Url, Config.Token, Debug, Config.Cert)
		values := make(map[string]string)
		values["project"] = Project
		values["cluster"] = Cluster
		values["virtual_cluster"] = VirtualCluster
		// values["application"] = Application
		if Role != "" {
			isSet, err_set := restClient.AddRoleToUser(Username, Role, values)
			if !isSet {
				return err_set
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(grantCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// grantCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// grantCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	grantCmd.Flags().StringVarP(&Username, "user", "", "", "Username required")
	grantCmd.MarkFlagRequired("user")
	grantCmd.Flags().StringVarP(&Role, "role", "", "", "Role required")
	grantCmd.MarkFlagRequired("role")
}
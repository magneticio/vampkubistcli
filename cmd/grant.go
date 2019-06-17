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
	"regexp"
	"strings"

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/spf13/cobra"
)

var Role string
var Permission string
var Kind string

// grantCmd represents the grant command
var grantCmd = &cobra.Command{
	Use:   "grant",
	Short: "Grant a role or permission to a user for an object",
	Long: AddAppName(`Usage:
$AppName grant --user user1 --role admin -p default
$AppName grant --user user1 --permission rwda -p project -c cluster -r virtualcluster -a application --kind deployment --name deploymentname
Permissions follow the format:
	r = read
	w = write
	d = delete
	a = edit access to resource for other users
`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		restClient := client.NewRestClient(Config.Url, Config.RefreshToken, Config.APIVersion, logging.Verbose, Config.Cert)
		values := make(map[string]string)
		values["project"] = Project
		values["cluster"] = Cluster
		values["virtual_cluster"] = VirtualCluster
		values["application"] = Application
		values[Kind] = Name
		if Role != "" {
			isSet, err_set := restClient.AddRoleToUser(Username, Role, values)
			if !isSet {
				return err_set
			}
			fmt.Println(Role + "is added to user " + Username)
		} else if Permission != "" {

			lowkPermission := strings.ToLower(Permission)

			// Regex matches all permutations of lowercase rwda with max length 4 and only once character per type
			reg := "^[rwda]{1,4}$"

			match, regexErr := regexp.MatchString(reg, lowkPermission)
			if regexErr != nil {
				return regexErr
			}
			validity := match &&
				strings.Count(lowkPermission, "r") <= 1 &&
				strings.Count(lowkPermission, "w") <= 1 &&
				strings.Count(lowkPermission, "d") <= 1 &&
				strings.Count(lowkPermission, "a") <= 1

			if !validity {
				return errors.New("Permission format is invalid")
			}

			isSet, err_set := restClient.UpdateUserPermission(Username, lowkPermission, values)
			if !isSet {
				return err_set
			}
			fmt.Println("Permission is added to user " + Username)
		} else {
			return errors.New("Resource to be granted is missing. Specify either permission or role.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(grantCmd)
	grantCmd.Flags().StringVarP(&Username, "user", "", "", "Username required")
	grantCmd.MarkFlagRequired("user")
	grantCmd.Flags().StringVarP(&Kind, "kind", "k", "", "")
	grantCmd.Flags().StringVarP(&Name, "name", "n", "", "")
	grantCmd.Flags().StringVarP(&Permission, "permission", "", "", "")
	grantCmd.Flags().StringVarP(&Role, "role", "", "", "")
}

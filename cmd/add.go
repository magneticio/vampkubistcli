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
	"strings"

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new user",
	Long: AddAppName(`Add a new user:
    $AppName add username
    This will print command to login for a new user.
  `),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("Not Enough Arguments")
		}
		Type = args[0]
		Name = args[1]

		if Type == "user" {
			Username := strings.ToLower(Name)
			// TODO: this is a temporary workaround it will be handled in the backend
			temporarayPassword := util.RandomString(50)
			SourceFileType := "json"
			Source := "{\"userName\":\"" + Username + "\",\"password\":\"" + temporarayPassword + "\"}"
			restClient := client.NewRestClient(Config.Url, Config.Token, Debug, Config.Cert)
			values := make(map[string]string)
			values["project"] = Config.Project
			values["cluster"] = Config.Cluster
			values["virtual_cluster"] = Config.VirtualCluster
			values["application"] = Application
			isCreated, createError := restClient.Create(Type, Name, Source, SourceFileType, values)
			if !isCreated {
				return createError
			}
			fmt.Printf("User created.\n")
			token, loginError := restClient.Login(Username, temporarayPassword)
			if loginError != nil {
				return loginError
			}
			fmt.Printf("User can login with:\n")
			fmt.Printf("vamp login --url %v --user %v --initial --token %v --cert <<EOF \"%v\"\nEOF\n", Config.Url, Username, token, Config.Cert)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

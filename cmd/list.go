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

	"github.com/magneticio/vamp2cli/client"
	"github.com/spf13/cobra"
)

var Detailed bool

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "lists resources",
	Long: AddAppName(`To list a resource type
Run as $AppName list resourceType

Example:
    $AppName list project
    $AppName list -p myproject cluster`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			Type = args[0]
		} else {
			return errors.New("Not Enough Arguments")
		}
		// fmt.Println("get called for type " + Type + " with name " + Name)
		restClient := client.NewRestClient(Config.Url, Config.Token, Debug, Config.Cert)
		values := make(map[string]string)
		values["project"] = Config.Project
		values["cluster"] = Config.Cluster
		values["virtual_cluster"] = Config.VirtualCluster
		values["application"] = Application
		result, err := restClient.List(Type, OutputType, values, !Detailed)
		if err == nil {
			if strings.TrimSuffix(result, "\n") != "[]" {
				fmt.Printf(result)
			}
			return nil
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	listCmd.Flags().StringVarP(&OutputType, "output", "o", "yaml", "Output format yaml or json")
	listCmd.Flags().BoolVarP(&Detailed, "detailed", "", false, "list detailed info")
}

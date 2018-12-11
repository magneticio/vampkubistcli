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

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a resource as yaml or json",
	Long: `To get a resource
Run as vamp2cli get resourceType ResourceName
Get show resource as yaml by defualt
By adding -o json, output can be converted to json

Example:
    vamp2cli get project myproject
    vamp2cli get -p myproject cluster mycluster
    vamp2cli get -p myproject cluster mycluster -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("Not Enough Arguments")
		}
		Type = args[0]
		Name = args[1]
		// fmt.Println("get called for type " + Type + " with name " + Name)
		restClient := client.NewRestClient(Config.Url, Config.Token, Debug)
		values := make(map[string]string)
		values["project"] = Config.Project
		values["cluster"] = Config.Cluster
		values["virtual_cluster"] = Config.VirtualCluster
		values["application"] = Application
		result, err := restClient.Get(Type, Name, OutputType, values)
		if err == nil {
			fmt.Printf(result)
			return nil
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// getCmd.Flags().StringVarP(&Name, "name", "n", "default", "Name Required")
	// getCmd.MarkFlagRequired("name")
	getCmd.Flags().StringVarP(&OutputType, "output", "o", "yaml", "Output format yaml or json")
}

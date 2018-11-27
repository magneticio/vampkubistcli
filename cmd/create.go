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
	"io/ioutil"

	"github.com/magneticio/vamp2cli/client"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a resource",
	Long: `To create a resource
Run as vamp2cli create resourceType ResourceName

Example:
    vamp2cli create project myproject
    vamp2cli create -p myproject project`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			return
		}
		Type = args[0]
		Name = args[1]
		// fmt.Println("create called for type " + Type + " with name " + Name)
		Source := SourceString
		if Source == "" {
			b, err := ioutil.ReadFile(SourceFile) // just pass the file name
			if err != nil {
				fmt.Print(err)
			}
			Source = string(b)
		}
		restClient := client.NewRestClient(Config.Url, Config.Token, Debug)
		values := make(map[string]string)
		values["project"] = Config.Project
		values["cluster"] = Config.Cluster
		values["virtual_cluster"] = Config.VirtualCluster
		values["application"] = Application
		isCreated, _ := restClient.Create(Type, Name, Source, SourceFileType, values)
		if !isCreated {
			fmt.Println("Not Created " + Type + " with name " + Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// createCmd.Flags().StringVarP(&Name, "name", "n", "default", "Name Required")
	// createCmd.MarkFlagRequired("name")
	createCmd.Flags().StringVarP(&SourceString, "string", "s", "", "Source from string")
	createCmd.Flags().StringVarP(&SourceFile, "file", "f", "", "Source from file")
	createCmd.Flags().StringVarP(&SourceFileType, "input", "i", "yaml", "Source file type yaml or json")

}

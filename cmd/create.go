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

var Type string
var Name string
var SourceString string
var SourceFile string

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			Type = args[0]
		}
		fmt.Println("create called for type " + Type + " with name " + Name)
		b, err := ioutil.ReadFile(SourceFile) // just pass the file name
		if err != nil {
			fmt.Print(err)
		}
		Source := string(b)
		if Type == "project" {
			restClient := client.NewRestClient(Config.Url, Config.Token)
			isCreated, _ := restClient.Create("projects", "project", Name, Source, "yaml")
			if !isCreated {
				fmt.Println("Not Created " + Type + " with name " + Name)
			}
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
	createCmd.Flags().StringVarP(&Name, "name", "n", "default", "Name Required")
	createCmd.MarkFlagRequired("name")
	createCmd.Flags().StringVarP(&SourceFile, "file", "f", "", "Source from file")

}

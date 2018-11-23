// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
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
		// fmt.Println("create called for type " + Type + " with name " + Name)
		b, err := ioutil.ReadFile(SourceFile) // just pass the file name
		if err != nil {
			fmt.Print(err)
		}
		Source := string(b)
		restClient := client.NewRestClient(Config.Url, Config.Token)
		values := make(map[string]string)
		values["project"] = Config.Project
		values["cluster"] = Config.Cluster
		values["virtual_cluster"] = Config.VirtualCluster
		isUpdated, _ := restClient.Update(Type, Name, Source, SourceFileType, values)
		if !isUpdated {
			fmt.Println("Not Updated " + Type + " with name " + Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	updateCmd.Flags().StringVarP(&Name, "name", "n", "default", "Name Required")
	updateCmd.MarkFlagRequired("name")
	updateCmd.Flags().StringVarP(&SourceFile, "file", "f", "", "Source from file")
	updateCmd.Flags().StringVarP(&SourceFileType, "input", "i", "yaml", "Source file type yaml or json")
}

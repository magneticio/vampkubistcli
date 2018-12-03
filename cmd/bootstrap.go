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
	"encoding/json"
	"fmt"

	b64 "encoding/base64"

	"github.com/magneticio/vamp2cli/client"
	"github.com/magneticio/vamp2cli/kubernetes"
	"github.com/spf13/cobra"
)

// bootstrapCmd represents the bootstrap command
var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			return
		}
		Type = args[0]
		Name = args[1]

		if Type == "cluster" {
			url, crt, token, err := kubeclient.BootstrapVampService()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			metadataMap := make(map[string]string)
			metadataMap["url"] = url
			metadataMap["cacertdata"] = b64.StdEncoding.EncodeToString([]byte(crt))
			metadataMap["serviceaccount_token"] = b64.StdEncoding.EncodeToString([]byte(token))
			metadata := &client.Metadata{Metadata: metadataMap}
			SourceRaw, err_marshall := json.Marshal(metadata)
			if err_marshall != nil {
				fmt.Printf("Error: %v", err_marshall)
			}
			Source := string(SourceRaw)
			// fmt.Printf("Source: %v", Source)
			SourceFileType = "json"
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
		}
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bootstrapCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bootstrapCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

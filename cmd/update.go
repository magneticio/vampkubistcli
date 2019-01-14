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
	"errors"

	"github.com/magneticio/vamp2cli/client"
	"github.com/magneticio/vamp2cli/util"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates a resource",
	Long: AddAppName(`To update a resource
Run as $AppName update resourceType ResourceName

Example:
    $AppName update project myproject -f project.yaml
    $AppName update -p myproject cluster mycluster -f cluster.yaml`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("Not Enough Arguments")
		}
		Type = args[0]
		Name = args[1]
		// fmt.Println("create called for type " + Type + " with name " + Name)
		Source := SourceString
		if Source == "" {
			b, err := util.UseSourceUrl(SourceFile) // just pass the file name
			if err != nil {
				// fmt.Print(err)
				return err
			}
			Source = string(b)
		}
		// This is a specific operation for vamp_service
		if client.ResourceTypeConversion(Type) == "vamp_service" && len(Hosts) > 0 {
			SourceJson, err := util.Convert(SourceFileType, "json", Source)
			if err != nil {
				return err
			}
			var vampService client.VampService
			err_json := json.Unmarshal([]byte(SourceJson), &vampService)
			if err_json != nil {
				return err_json
			}
			vampService.Hosts = append(Hosts, vampService.Hosts...)
			SourceRaw, err := json.Marshal(vampService)
			if err != nil {
				return err
			}
			Source = string(SourceRaw)
			// fmt.Println(Source)
			SourceFileType = "json"
		}
		restClient := client.NewRestClient(Config.Url, Config.Token, Debug, Config.Cert)
		values := make(map[string]string)
		values["project"] = Config.Project
		values["cluster"] = Config.Cluster
		values["virtual_cluster"] = Config.VirtualCluster
		values["application"] = Application
		isUpdated, err_update := restClient.Update(Type, Name, Source, SourceFileType, values)
		if !isUpdated {
			return err_update
		}
		return nil
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
	// updateCmd.Flags().StringVarP(&Name, "name", "n", "default", "Name Required")
	// updateCmd.MarkFlagRequired("name")
	updateCmd.Flags().StringVarP(&SourceString, "string", "s", "", "Source from string")
	updateCmd.Flags().StringVarP(&SourceFile, "file", "f", "", "Source from file")
	updateCmd.Flags().StringVarP(&SourceFileType, "input", "i", "yaml", "Source file type yaml or json")
	updateCmd.Flags().StringSliceVarP(&Hosts, "host", "", []string{}, "host to add to vamp service, Comma separated lists are supported")

}

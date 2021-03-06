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
	"fmt"

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/magneticio/vampkubistcli/models"
	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
)

var Init bool

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a resource",
	Long: AddAppName(`To create a resource
Run as $AppName create resourceType ResourceName

Example:
    $AppName create project myproject -f project.yaml
    $AppName create -p myproject cluster mycluster -f cluster.yaml`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			// fmt.Errorf("Not enough arguments\n")
			return errors.New("Not enough arguments")
		}
		Type = args[0]
		Name = args[1]
		Source := SourceString
		if Init {
			Source = "{}"
			SourceFileType = "json"
		}
		if Source == "" {
			b, err := util.UseSourceUrl(SourceFile) // just pass the file name
			if err != nil {
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
			var vampService models.VampService
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

			SourceFileType = "json"
		}

		restClient := client.NewRestClient(Config.Url, Config.Token, Config.APIVersion, logging.Verbose, Config.Cert, &TokenStore)
		values := make(map[string]string)
		values["project"] = Config.Project
		values["cluster"] = Config.Cluster
		values["virtual_cluster"] = Config.VirtualCluster
		values["application"] = Application
		values["destination"] = Destination
		values["experiment"] = Experiment
		values["port"] = Port
		values["subset"] = Subset
		isCreated, createError := restClient.Create(Type, Name, Source, SourceFileType, values)
		if !isCreated {
			return createError
		}
		fmt.Println(Type + " " + Name + " is created")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&SourceString, "string", "s", "", "Source from string")
	createCmd.Flags().StringVarP(&SourceFile, "file", "f", "", "Source from file")
	createCmd.Flags().StringVarP(&SourceFileType, "input", "i", "yaml", "Source file type yaml or json")
	createCmd.Flags().BoolVarP(&Init, "init", "", false, "initialize as empty source")
	createCmd.Flags().StringVarP(&Destination, "destination", "", "", "destination name for metrics")
	createCmd.Flags().StringVarP(&Experiment, "experiment", "", "", "experiment name for metrics")
	createCmd.Flags().StringVarP(&Port, "port", "", "", "port number for metrics")
	createCmd.Flags().StringVarP(&Subset, "subset", "", "", "subset name for metrics")
	createCmd.Flags().StringSliceVarP(&Hosts, "host", "", []string{}, "host to add to vamp service, Comma separated lists are supported")

}

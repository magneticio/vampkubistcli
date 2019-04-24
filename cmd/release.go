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
	"github.com/spf13/cobra"
)

var Api string
var Subset string
var Port string
var Destination string
var SubsetLabels map[string]string

// releaseCmd represents the release command
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release a new subset with labels",
	Long: AddAppName(`eg.:
$AppName release shop-vamp-service --destination shop-destination --port port --subset subset2 -l version=version2`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Not Enough Arguments")
		}
		Type := "canary_release"
		VampService := args[0]
		Name := VampService + "-" + Destination + "-" + Subset

		// fmt.Printf("%v %v %v\n", Type, Name, SubsetLabels)
		canaryRelease := models.CanaryRelease{
			VampService:  VampService,
			Destination:  Destination,
			Port:         Port,
			Subset:       Subset,
			SubsetLabels: SubsetLabels,
		}
		SourceRaw, marshallError := json.Marshal(canaryRelease)
		if marshallError != nil {
			return marshallError
		}
		Source := string(SourceRaw)
		SourceFileType = "json"

		restClient := client.NewRestClient(Config.Url, Config.Token, Config.APIVersion, logging.Verbose, Config.Cert)
		values := make(map[string]string)
		values["project"] = Config.Project
		values["cluster"] = Config.Cluster
		values["virtual_cluster"] = Config.VirtualCluster
		values["application"] = Application
		isCreated, createError := restClient.Create(Type, Name, Source, SourceFileType, values)
		if !isCreated {
			return createError
		}
		fmt.Println(Type + " " + Name + " is created")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)

	releaseCmd.Flags().StringVarP(&Destination, "destination", "", "", "Destination to use in the release")
	releaseCmd.Flags().StringVarP(&Port, "port", "", "", "Port to use in the release")
	releaseCmd.Flags().StringVarP(&Subset, "subset", "", "", "Subset to use in the release")
	releaseCmd.Flags().StringToStringVarP(&SubsetLabels, "label", "l", map[string]string{}, "Subset labels, multiple labels are allowed")

}

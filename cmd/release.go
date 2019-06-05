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
	"strconv"

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/magneticio/vampkubistcli/models"
	"github.com/spf13/cobra"
)

var Api string
var Subset string
var Port string
var Period string
var Step string
var Destination string
var SubsetLabels map[string]string
var ReleaseType string

// releaseCmd represents the release command
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release a new subset with labels",
	Long: AddAppName(`eg.:
$AppName release shop-vamp-service --destination shop-destination --port port --subset subset2 -l version=version2 --type time`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Not Enough Arguments")
		}
		Type := "canary_release"
		VampService := args[0]

		policies := []models.PolicyReference{}

		allowedReleaseTypes := map[string]string{"time": "TimedCanaryReleasePolicy", "health": "HealthBasedCanaryReleasePolicy"}

		if ReleaseType != "" {

			logging.Info("Release type is %v", allowedReleaseTypes[ReleaseType])

			if allowedReleaseTypes[ReleaseType] == "" {
				return errors.New("Release type is not valid")
			}

			policies = []models.PolicyReference{models.PolicyReference{
				Name: allowedReleaseTypes[ReleaseType],
			}}

		}

		var portReference *int

		if Port != "" {
			portInt, convErr := strconv.Atoi(Port)
			if convErr != nil {
				return convErr
			}
			portReference = &portInt
		}

		var periodReference *int

		if Period != "" {
			periodInt, convErr := strconv.Atoi(Period)
			if convErr != nil {
				return convErr
			}
			periodReference = &periodInt
		}

		var stepReference *int

		if Step != "" {
			stepInt, convErr := strconv.Atoi(Step)
			if convErr != nil {
				return convErr
			}
			stepReference = &stepInt
		}

		// fmt.Printf("%v %v %v\n", Type, Name, SubsetLabels)
		canaryRelease := models.CanaryRelease{
			VampService:  VampService,
			Destination:  Destination,
			Port:         portReference,
			UpdatePeriod: periodReference,
			UpdateStep:   stepReference,
			Subset:       Subset,
			SubsetLabels: SubsetLabels,
			Policies:     policies,
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
		isCreated, createError := restClient.Create(Type, VampService, Source, SourceFileType, values)
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
	releaseCmd.Flags().StringVarP(&Period, "period", "", "", "Period of updates")
	releaseCmd.Flags().StringVarP(&Step, "step", "", "", "Step is the percetage change at each step")
	releaseCmd.Flags().StringVarP(&Subset, "subset", "", "", "Subset to use in the release")
	releaseCmd.Flags().StringVarP(&ReleaseType, "type", "", "", "Type of canary release to use")
	releaseCmd.Flags().StringToStringVarP(&SubsetLabels, "label", "l", map[string]string{}, "Subset labels, multiple labels are allowed")

}

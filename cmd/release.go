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
	"github.com/spf13/cobra"
)

var Subset string
var Destination string
var SubsetLabels map[string]string

// releaseCmd represents the release command
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release a new subset with labels",
	Long: AddAppName(`eg.:
$AppName release shop-vamp-service --destination shop-destination --subset subset2 -l version=version2`),
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
		canaryRelease := client.CanaryRelease{
			VampService:  VampService,
			Destination:  Destination,
			Subset:       Subset,
			SubsetLabels: SubsetLabels,
		}
		SourceRaw, err_marshall := json.Marshal(canaryRelease)
		if err_marshall != nil {
			// fmt.Printf("Error: %v", err_marshall)
			return err_marshall
		}
		Source := string(SourceRaw)
		// fmt.Printf("Source: %v", Source)
		SourceFileType = "json"
		restClient := client.NewRestClient(Config.Url, Config.Token, Debug, Config.Cert)
		values := make(map[string]string)
		values["project"] = Config.Project
		values["cluster"] = Config.Cluster
		values["virtual_cluster"] = Config.VirtualCluster
		values["application"] = Application
		isCreated, err_create := restClient.Create(Type, Name, Source, SourceFileType, values)
		if !isCreated {
			// fmt.Println("Not Created " + Type + " with name " + Name)
			return err_create
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// releaseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// releaseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	releaseCmd.Flags().StringVarP(&Destination, "destination", "", "", "Destination to use in the release")
	releaseCmd.Flags().StringVarP(&Subset, "subset", "", "", "Subset to use in the release")
	releaseCmd.Flags().StringToStringVarP(&SubsetLabels, "label", "l", map[string]string{}, "Subset labels, multiple labels are allowed")

}

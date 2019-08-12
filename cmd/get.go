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
	"time"

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
)

var JsonPath string
var WaitUntilAvailable bool
var NumberOfTrialLimit int

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a resource as yaml or json",
	Long: AddAppName(`To get a resource
Run as $AppName get resourceType ResourceName
Get show resource as yaml by defualt
By adding -o json, output can be converted to json

Example:
    $AppName get project myproject
    $AppName get -p myproject cluster mycluster
    $AppName get -p myproject cluster mycluster -o json

Json path example with wait
    $AppName get gateway shop-gateway -o=json --jsonpath '$.status.ip' --wait
    `),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("Not Enough Arguments")
		}
		Type = args[0]
		Name = args[1]
		// fmt.Println("get called for type " + Type + " with name " + Name)
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
		first := true
		var result string
		var getError error
		numberOfTrials := 0
		for (WaitUntilAvailable || first) && (numberOfTrials < NumberOfTrialLimit || NumberOfTrialLimit == 0) {
			first = false
			result, getError = restClient.Get(Type, Name, OutputType, values)
			numberOfTrials++
			if getError == nil {
				if JsonPath != "" {
					resultPath, jsonpathError := util.GetJsonPath(result, OutputType, JsonPath)
					if jsonpathError != nil {
						getError = jsonpathError
						time.Sleep(5 * time.Second)
						continue
					}
					result = resultPath
				}
				if result != "" {
					fmt.Printf(result)
					return nil
				}
			} else {
				time.Sleep(10 * time.Second)
			}
		}
		if getError != nil {
			return getError
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringVarP(&OutputType, "output", "o", "yaml", "Output format yaml or json")
	getCmd.Flags().StringVarP(&JsonPath, "jsonpath", "", "", "Json path to access specific parts of the object")
	getCmd.Flags().BoolVarP(&WaitUntilAvailable, "wait", "w", false, "Wait until output is available")
	getCmd.Flags().IntVarP(&NumberOfTrialLimit, "number-of-tries", "", 0, "Number of Tries when failed, this flag should be used with wait flag (0 is infinite)")
	getCmd.Flags().StringVarP(&Destination, "destination", "", "", "destination name for metrics")
	getCmd.Flags().StringVarP(&Experiment, "experiment", "", "", "experiment name for metrics")
	getCmd.Flags().StringVarP(&Port, "port", "", "", "port number for metrics")
	getCmd.Flags().StringVarP(&Subset, "subset", "", "", "subset name for metrics")
}

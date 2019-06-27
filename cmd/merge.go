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

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/config"
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
)

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merges a resource",
	Long: config.AddAppName(`To merge a resource
Run as $AppName merge resourceType ResourceName

Example:
    $AppName merge project myproject -f project.yaml
    $AppName merge -p myproject cluster mycluster -f cluster.yaml`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("Not Enough Arguments")
		}
		Type = args[0]
		Name = args[1]
		source := SourceString
		if source == "" {
			b, err := util.UseSourceUrl(SourceFile) // just pass the file name
			if err != nil {
				return err
			}
			source = string(b)
		}
		restClient := client.ClientFromConfig(&config.Config, logging.Verbose)
		values := make(map[string]string)
		values["project"] = config.Config.Project
		values["cluster"] = config.Config.Cluster
		values["virtual_cluster"] = config.Config.VirtualCluster
		values["application"] = Application
		spec, getSpecError := restClient.GetSpec(Type, Name, SourceFileType, values)
		if getSpecError != nil {
			return getSpecError
		}

		updatedSource, mergeError := util.Merge(spec, source, SourceFileType)
		if mergeError != nil {
			return mergeError
		}

		isUpdated, updateError := restClient.Update(Type, Name, updatedSource, SourceFileType, values)
		if !isUpdated {
			return updateError
		}
		fmt.Println(Type + " " + Name + " is merged")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().StringVarP(&SourceString, "string", "s", "", "Source from string")
	mergeCmd.Flags().StringVarP(&SourceFile, "file", "f", "", "Source from file")
	mergeCmd.Flags().StringVarP(&SourceFileType, "input", "i", "yaml", "Resource file type yaml or json")
}

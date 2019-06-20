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
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a resource",
	Long: config.AddAppName(`To delete a resource
Run as $AppName delete resourceType ResourceName

Example:
    $AppName delete project myproject
    $AppName delete -p myproject cluster mycluster`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("Not Enough Arguments")
		}
		Type = args[0]
		Name = args[1]
		// fmt.Println("delete called for type " + Type + " with name " + Name)
		restClient := client.ClientFromConfig(&config.Config, logging.Verbose)
		values := make(map[string]string)
		values["project"] = config.Config.Project
		values["cluster"] = config.Config.Cluster
		values["virtual_cluster"] = config.Config.VirtualCluster
		values["application"] = Application
		isDeleted, deleteError := restClient.Delete(Type, Name, values)
		if !isDeleted {
			return deleteError
		}
		fmt.Println(Type + " " + Name + " is deleted")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

}

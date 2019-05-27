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

	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set, Get , Edit configuration of client",
	Long: AddAppName(`To get all configuration parameters:
  $AppName config get
To set configuration parameters:
  $AppName config set -p myproject -c mycluster`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Not Enough Arguments, use get or set")
		}
		function := args[0]
		if function == "set" {
			writeConfigError := WriteConfigFile()
			if writeConfigError != nil {
				return writeConfigError
			}
		} else if function == "get" {
			SourceRaw, marshalError := json.Marshal(Config)
			if marshalError != nil {
				return marshalError
			}
			result, convertError := util.Convert("json", OutputType, string(SourceRaw))
			if convertError != nil {
				return convertError
			}
			fmt.Printf("%v", result)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().StringVarP(&OutputType, "output", "o", "yaml", "Output format yaml or json")

}

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

	"github.com/magneticio/vamp2cli/util"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set, Get , Edit configuration of client",
	Long: `To get all configuration parameters:
  vamp2cli config get
To set configuration parameters:
  vamp2cli config set -p myproject -c mycluster`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Not Enough Arguments, use get or set")
		}
		function := args[0]
		if function == "set" {
			WriteConfigFile()
		} else if function == "get" {
			SourceRaw, err_marshall := json.Marshal(Config)
			if err_marshall != nil {
				return err_marshall
			}
			result, err_convert := util.Convert("json", OutputType, string(SourceRaw))
			if err_convert != nil {
				return err_convert
			}
			fmt.Printf("%v", result)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	configCmd.Flags().StringVarP(&OutputType, "output", "o", "yaml", "Output format yaml or json")

}
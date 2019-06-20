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
	"github.com/magneticio/vampkubistcli/config"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set global project, cluster and virtual cluster",
	Long: config.AddAppName(`When you need to use the command line client for longer periods,
  it is cumbersome to set project, cluster and virtualcluster in every command.
  You can set these variables with a set command
  and it is stored in global configuration.
  Example:
  $AppName set -p myproject -c mycluster -v myvirtualcluster

  Please use $AppName config set instead of this method.
  `),
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		config.WriteConfigFile()
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}

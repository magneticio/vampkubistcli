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
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:           "version",
	Short:         config.AddAppName("Print the version number of $AppName"),
	Long:          config.AddAppName(`All software has versions. This is $AppName's`),
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) >= 1 {
			function := args[0]
			if function == "clean" {
				fmt.Printf("%v\n", Version)
			}
		} else {
			fmt.Printf("Version: %v\n", Version)
			fmt.Printf("Backend Version: %v\n", BackendVersion)
		}

	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

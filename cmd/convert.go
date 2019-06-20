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

	"github.com/magneticio/vampkubistcli/config"
	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "convert json to yaml and yam to json",
	Long: config.AddAppName(`Convert is an utility method for easy convertion between json and yaml.
    $AppName supports both json and yaml but if you like to convert them.
    Example converting json to yaml:
    $AppName convert -i json -f filepath.json
    This will print yaml version of the json object.
  `),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		source := SourceString
		if source == "" {
			b, err := util.UseSourceUrl(SourceFile) // just pass the file name
			if err != nil {
				return err
			}
			source = string(b)
		}
		result, err := util.Convert(SourceFileType, OutputType, source)
		if err == nil {
			fmt.Printf(result)
			return nil
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringVarP(&SourceString, "string", "s", "", "Source from string")
	convertCmd.Flags().StringVarP(&SourceFile, "file", "f", "", "Source from file")
	convertCmd.Flags().StringVarP(&SourceFileType, "input", "i", "yaml", "Source file type yaml or json")
	convertCmd.Flags().StringVarP(&OutputType, "output", "o", "yaml", "Output format yaml or json")
}

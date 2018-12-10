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
	"io/ioutil"

	"github.com/magneticio/vamp2cli/client"
	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "convert json to yaml and yam to json",
	Long: `Convert is an utility method for easy convertion between json and yaml.
    vamp2cli supports both json and yaml but if you like to convert them.
    Example converting json to yaml:
    vamp2cli convert -i json -f filepath.json
    This will print yaml version of the json object.
  `,
	Run: func(cmd *cobra.Command, args []string) {
		Source := SourceString
		if Source == "" {
			b, err := ioutil.ReadFile(SourceFile) // just pass the file name
			if err != nil {
				fmt.Print(err)
			}
			Source = string(b)
		}
		restClient := client.NewRestClient(Config.Url, Config.Token, Debug)
		result, err := restClient.Convert(SourceFileType, OutputType, Source)
		if err == nil {
			fmt.Printf(result)
		}
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// convertCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// convertCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	convertCmd.Flags().StringVarP(&SourceString, "string", "s", "", "Source from string")
	convertCmd.Flags().StringVarP(&SourceFile, "file", "f", "", "Source from file")
	convertCmd.Flags().StringVarP(&SourceFileType, "input", "i", "yaml", "Source file type yaml or json")
	convertCmd.Flags().StringVarP(&OutputType, "output", "o", "yaml", "Output format yaml or json")
}

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
	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
)

var URL string
var Path string

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a resource from url",
	Long: AddAppName(`Utility method for downloading resources:
    eg:.
    $AppName download --url URL --path path-of-file`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := util.DownloadFile(Path, URL)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVarP(&URL, "url", "", "", "URL to download")
	downloadCmd.MarkFlagRequired("url")
	downloadCmd.Flags().StringVarP(&Path, "path", "", "", "Path to write the file")
	downloadCmd.MarkFlagRequired("path")
}

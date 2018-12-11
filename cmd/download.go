// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"github.com/magneticio/vamp2cli/util"
	"github.com/spf13/cobra"
)

var URL string
var Path string

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a resource from url",
	Long: `Utility method for downloading resources:
eg:.
vamp2cli download --url URL --path path-of-file`,
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	downloadCmd.Flags().StringVarP(&URL, "url", "", "", "URL to download")
	downloadCmd.MarkFlagRequired("url")
	downloadCmd.Flags().StringVarP(&Path, "path", "", "", "Path to write the file")
	downloadCmd.MarkFlagRequired("path")
}
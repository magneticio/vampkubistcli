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
	"io/ioutil"
	"strings"

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var userConfigFilePath string

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new user",
	Long: AddAppName(`Add a new user:
    $AppName add user username
    This will print command to login for a new user.
    It is also possible to generate login configuration file for added user:
    $AppName add user username --user-config-output-path ./configurationFile.yaml
  `),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("Not Enough Arguments")
		}
		Type = args[0]
		Name = args[1]

		if Type == "user" {
			Username := strings.ToLower(Name)
			// TODO: this is a temporary workaround it will be handled in the backend
			temporarayPassword := util.RandomString(50)
			SourceFileType := "json"
			Source := "{\"userName\":\"" + Username + "\",\"password\":\"" + temporarayPassword + "\"}"
			restClient := client.NewRestClient(Config.Url, Config.Token, Config.APIVersion, logging.Verbose, Config.Cert)
			values := make(map[string]string)
			values["project"] = Config.Project
			values["cluster"] = Config.Cluster
			values["virtual_cluster"] = Config.VirtualCluster
			values["application"] = Application
			isCreated, createError := restClient.Create(Type, Name, Source, SourceFileType, values)
			if !isCreated {
				return createError
			}
			fmt.Printf("User created.\n")
			token, loginError := restClient.Login(Username, temporarayPassword)
			if loginError != nil {
				return loginError
			}
			fmt.Printf("User can login with:\n")
			fmt.Printf("%v login --url %v --user %v --initial --token %v --cert <<EOF \"%v\"\nEOF\n", AppName, Config.Url, Username, token, Config.Cert)

			// Write the file is called after printing the output to handle avoid file write errors blocking user creation
			if userConfigFilePath != "" {
				userConfig := &config{
					Url:      Config.Url,
					Cert:     Config.Cert,
					Username: Username,
					Token:    token,
				}
				writeConfigError := writeConfigToFile(userConfig, userConfigFilePath)
				if writeConfigError != nil {
					return writeConfigError
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&userConfigFilePath, "user-config-output-path", "", "", "Generated user configuration file output path. Path should be in an existing folder.")
}

func writeConfigToFile(userConfig *config, filename string) error {
	bs, marshallError := yaml.Marshal(userConfig)
	if marshallError != nil {
		return marshallError
	}
	fileWriteError := ioutil.WriteFile(filename, bs, 0644)
	if fileWriteError != nil {
		return fileWriteError
	}
	return nil
}

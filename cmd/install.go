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
	"fmt"
	"io/ioutil"

	"github.com/magneticio/vampkubistcli/kubernetes"
	"github.com/magneticio/vampkubistcli/models"
	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
)

var installConfigPath string
var configFileType string
var certFileName string

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Vamp Management in your cluster",
	Long: `
Example:
vamp install --configuration installconfig.yml --cert ./certiticate.crt

Configuration file example as yaml:
rootPassword: root
databaseUrl:
repoUsername: dockerhubusername
repoPassword: dockerhubpassword
vampVersion: 0.7.0

Leave databaseUrl empty to deploy an internal one
  `,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		configBtye, readErr := util.UseSourceUrl(installConfigPath) // just pass the file name
		if readErr != nil {
			return readErr
		}
		source := string(configBtye)
		configJson, convertErr := util.Convert(configFileType, "json", source)
		if convertErr != nil {
			return convertErr
		}
		var config models.VampConfig
		unmarshallError := json.Unmarshal([]byte(configJson), &config)
		if unmarshallError != nil {
			return unmarshallError
		}
		validatedConfig, configError := kubeclient.VampConfigValidateAndSetupDefaults(&config)
		if configError != nil {
			return configError
		}
		fmt.Printf("Vamp Configuration validated.\n")
		url, cert, _, err := kubeclient.InstallVampService(validatedConfig)
		if err != nil {
			return err
		}
		writeError := ioutil.WriteFile(certFileName, cert, 0644)
		if writeError != nil {
			return writeError
		}
		fmt.Printf("Vamp Service Installed.\n")
		fmt.Printf("Login with:\n")
		fmt.Printf("vamp login --url %v --user root --cert %v\n", url, certFileName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	installCmd.Flags().StringVarP(&installConfigPath, "configuration", "", "", "Installation configuration file path")
	installCmd.MarkFlagRequired("configuration")
	installCmd.Flags().StringVarP(&configFileType, "input", "i", "yaml", "Configuration file type yaml or json")
	installCmd.Flags().StringVarP(&certFileName, "cert", "", "cert.crt", "Certificate file output path")
}

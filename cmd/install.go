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

	"github.com/magneticio/vampkubistcli/config"
	"github.com/magneticio/vampkubistcli/kubernetes"
	"github.com/magneticio/vampkubistcli/models"
	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var installConfigPath string
var configFileType string
var certFileName string

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Vamp Management in your cluster",
	Long: config.AddAppName(`
Example:
$AppName install --configuration installconfig.yml --certificate-output-path ./certiticate.crt

Configuration file example as yaml:
rootPassword: root1234
databaseUrl:
databaseName: vamp
imageName: magneticio/vampkubist
repoUsername: dockerHubUsername
repoPassword: dockerHubPassword
imageTag: 0.7.7
mode: IN_CLUSTER

Leave databaseUrl empty to deploy an internal one.

Install will generate certificates for the cluster which will be written to the certificate output path.
Install command is reentrant, it is possible to update the cluster with re-running the command.
  `),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if kubeConfigPath == "" {
			kubeConfigPath = viper.GetString("kubeconfig")
		}
		configBtye, readErr := util.UseSourceUrl(installConfigPath) // just pass the file name
		if readErr != nil {
			return readErr
		}
		source := string(configBtye)
		configJson, convertErr := util.Convert(configFileType, "json", source)
		if convertErr != nil {
			return convertErr
		}
		var cfg models.VampConfig
		unmarshallError := json.Unmarshal([]byte(configJson), &cfg)
		if unmarshallError != nil {
			return unmarshallError
		}
		validatedConfig, configError := kubeclient.VampConfigValidateAndSetupDefaults(&cfg)
		if configError != nil {
			return configError
		}
		fmt.Printf("Vamp Configuration validated.\n")
		url, cert, _, err := kubeclient.InstallVampService(validatedConfig, kubeConfigPath)
		if err != nil {
			return err
		}
		writeError := ioutil.WriteFile(certFileName, cert, 0644)
		if writeError != nil {
			return writeError
		}
		fmt.Printf("Vamp Service Installed.\n")
		fmt.Printf("Login with:\n")
		fmt.Printf("%v login --url %v --user root --cert %v\n", config.AppName, url, certFileName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().StringVarP(&installConfigPath, "configuration", "", "", "Installation configuration file path")
	installCmd.MarkFlagRequired("configuration")
	installCmd.Flags().StringVarP(&configFileType, "input", "i", "yaml", "Configuration file type yaml or json")
	installCmd.Flags().StringVarP(&certFileName, "certificate-output-path", "", "certificate.crt", "Certificate file output path")
	installCmd.Flags().StringVarP(&kubeConfigPath, "kubeconfig", "", "", "Kube Config path")
	viper.BindEnv("kubeconfig", "KUBECONFIG")
}

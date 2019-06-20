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

	b64 "encoding/base64"

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/config"
	"github.com/magneticio/vampkubistcli/kubernetes"
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/magneticio/vampkubistcli/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// bootstrapCmd represents the bootstrap command
var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "bootstrap a resource if it is applicable",
	Long: config.AddAppName(`Bootstrap is currently only supported for clusters
    When you want to bootstrap a new cluster with vamp.
    Make sure kube config setup and active locally.
    Then run:
    $AppName bootstrap cluster mycluster
    This will automacially read configuration and create vamp user in your cluster and
    make required set up in vamp. You can access the cluster with name mycluster.
  `),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("Not Enough Arguments")
		}
		Type = args[0]
		Name = args[1]

		if kubeConfigPath == "" {
			kubeConfigPath = viper.GetString("kubeconfig")
		}

		if Type == "cluster" {
			url, crt, token, err := kubeclient.BootstrapVampService(kubeConfigPath)
			if err != nil {
				// fmt.Printf("Error: %v\n", err)
				return err
			}
			metadataMap := make(map[string]string)
			metadataMap["url"] = url
			metadataMap["cacertdata"] = b64.StdEncoding.EncodeToString([]byte(crt))
			metadataMap["serviceaccount_token"] = b64.StdEncoding.EncodeToString([]byte(token))
			metadata := &models.Metadata{Metadata: metadataMap}
			SourceRaw, err_marshall := json.Marshal(metadata)
			if err_marshall != nil {
				return err_marshall
			}
			source := string(SourceRaw)
			SourceFileType = "json"
			restClient := client.ClientFromConfig(&config.Config, logging.Verbose)
			values := make(map[string]string)
			values["project"] = config.Config.Project
			values["cluster"] = config.Config.Cluster
			values["virtual_cluster"] = config.Config.VirtualCluster
			values["application"] = Application
			isCreated, err_create := restClient.Create(Type, Name, source, SourceFileType, values)
			if !isCreated {
				return err_create
			}
			return nil
		}
		return errors.New("Unsupported Type")
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)

	bootstrapCmd.Flags().StringVarP(&kubeConfigPath, "kubeconfig", "", "", "Kube Config path")
	viper.BindEnv("kubeconfig", "KUBECONFIG")
}

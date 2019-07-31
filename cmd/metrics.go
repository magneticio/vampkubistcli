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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/magneticio/vampkubistcli/kubernetes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var namespace string

// bootstrapCmd represents the bootstrap command
var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "get metrics",
	Long: AddAppName(`Get pods metrics for a given namespace

Example:
    $AppName metrics
  `),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if kubeConfigPath == "" {
			kubeConfigPath = viper.GetString("kubeconfig")
		}

		var pods kubeclient.PodMetricsList
		err := kubeclient.Metrics(kubeConfigPath, namespace, &pods)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}

		js, err := json.Marshal(pods)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}

		if OutputType == "yaml" {
			yaml, err := yaml.JSONToYAML(js)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return err
			}
			fmt.Print(string(yaml))
		} else {
			var prettyJSON bytes.Buffer
			error := json.Indent(&prettyJSON, js, "", "    ")
			if error != nil {
				fmt.Printf("Error: %v\n", err)
				return err
			}
			fmt.Print(string(prettyJSON.Bytes()))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(metricsCmd)

	metricsCmd.Flags().StringVarP(&namespace, "namespace", "", "vamp-system", "Namespace")
	metricsCmd.Flags().StringVarP(&kubeConfigPath, "kubeconfig", "", "", "Kube Config path")
	metricsCmd.Flags().StringVarP(&OutputType, "output", "o", "yaml", "Output format yaml or json")
	viper.BindEnv("kubeconfig", "KUBECONFIG")
}

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
var metricsKind string

// bootstrapCmd represents the bootstrap command
var k8sMetricsCmd = &cobra.Command{
	Use:   "k8smetrics",
	Short: "get k8s metrics",
	Long: AddAppName(`Get k8s pods metrics for a given namespace

Example:
    $AppName k8smetrics
  `),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if kubeConfigPath == "" {
			kubeConfigPath = viper.GetString("kubeconfig")
		}

		var pods kubeclient.PodMetricsList
		var err error
		var avgMetrics []kubeclient.PodAverageMetrics
		var podMetrics []kubeclient.PodContainersMetrics
		switch metricsKind {
		case "processed":
			err = kubeclient.GetProcessedMetrics(kubeConfigPath, namespace, &pods)
		case "average":
			avgMetrics, err = kubeclient.GetAverageMetrics(kubeConfigPath, namespace)
		case "simple":
			podMetrics, err = kubeclient.GetSimpleMetrics(kubeConfigPath, namespace)
		default:
			return fmt.Errorf(`Bad metrics kind "%v"`, metricsKind)
		}

		if err != nil {
			return err
		}

		var js []byte

		switch metricsKind {
		case "processed":
			js, err = json.Marshal(pods)
		case "average":
			js, err = json.Marshal(avgMetrics)
		case "simple":
			js, err = json.Marshal(podMetrics)
		default:
			return fmt.Errorf(`Bad metrics kind "%v"`, metricsKind)
		}

		if err != nil {
			return err
		}

		if OutputType == "yaml" {
			yaml, err := yaml.JSONToYAML(js)
			if err != nil {
				return err
			}
			fmt.Print(string(yaml))
		} else {
			var prettyJSON bytes.Buffer
			error := json.Indent(&prettyJSON, js, "", "    ")
			if error != nil {
				return err
			}
			fmt.Print(string(prettyJSON.Bytes()))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(k8sMetricsCmd)

	k8sMetricsCmd.Flags().StringVarP(&namespace, "namespace", "", "vamp-system", "Namespace")
	k8sMetricsCmd.Flags().StringVarP(&kubeConfigPath, "kubeconfig", "", "", "Kube Config path")
	k8sMetricsCmd.Flags().StringVarP(&OutputType, "output", "o", "yaml", "Output format yaml or json")
	k8sMetricsCmd.Flags().StringVarP(&metricsKind, "kind", "k", "simple", "Kind of metrics, simple, processed or average")
	viper.BindEnv("kubeconfig", "KUBECONFIG")
}

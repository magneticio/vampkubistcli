// Copyright © 2019 Developer <developer@vamp.io>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/magneticio/forklift/logging"
	"github.com/magneticio/forklift/util"
	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/models"
	"github.com/spf13/cobra"
)

var pushMetricCmd = &cobra.Command{
	Use:   "metric",
	Short: "Push a metric",
	Long: AddAppName(`To update a resource
	Run as $AppName update resourceType ResourceName

	Example:
    $AppName push metric mymetric --destination mydestination -f metric.yaml
    $AppName push metric mymetric --destination mydestination --port 8080 --subset mysubset -f metric.yaml`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) < 1 {
			return errors.New("Not Enough Arguments")
		}
		Type = "metric"
		Name = args[0]
		Source := SourceString
		if Source == "" {
			b, err := util.UseSourceUrl(SourceFile) // just pass the file name
			if err != nil {
				return err
			}
			Source = string(b)
		}
		// This is a specific operation for vamp_service
		if client.ResourceTypeConversion(Type) == "vamp_service" && len(Hosts) > 0 {
			SourceJson, err := util.Convert(SourceFileType, "json", Source)
			if err != nil {
				return err
			}
			var vampService models.VampService
			err_json := json.Unmarshal([]byte(SourceJson), &vampService)
			if err_json != nil {
				return err_json
			}
			vampService.Hosts = append(Hosts, vampService.Hosts...)
			SourceRaw, err := json.Marshal(vampService)
			if err != nil {
				return err
			}
			Source = string(SourceRaw)
			// fmt.Println(Source)
			SourceFileType = "json"
		}
		restClient := client.NewRestClient(Config.Url, Config.Token, Config.APIVersion, logging.Verbose, Config.Cert, &TokenStore)
		values := make(map[string]string)
		values["project"] = Config.Project
		values["cluster"] = Config.Cluster
		values["virtual_cluster"] = Config.VirtualCluster
		values["application"] = Application
		values["destination"] = Destination
		values["experiment"] = Experiment
		values["port"] = Port
		values["subset"] = Subset
		isUpdated, updateError := restClient.PushMetricValueInternal(Name, Source, SourceFileType, values)
		if !isUpdated {
			return updateError
		}
		fmt.Println(Type + " " + Name + " is updated")
		return nil
	},
}

func init() {
	pushCmd.AddCommand(pushMetricCmd)

	pushMetricCmd.Flags().StringVarP(&SourceString, "string", "s", "", "Source from string")
	pushMetricCmd.Flags().StringVarP(&SourceFile, "file", "f", "", "Source from file")
	pushMetricCmd.Flags().StringVarP(&SourceFileType, "input", "i", "yaml", "Source file type yaml or json")
	pushMetricCmd.Flags().StringVarP(&Destination, "destination", "", "", "destination name for metrics")
	pushMetricCmd.Flags().StringVarP(&Experiment, "experiment", "", "", "experiment name for metrics")
	pushMetricCmd.Flags().StringVarP(&Port, "port", "", "", "port number for metrics")
	pushMetricCmd.Flags().StringVarP(&Subset, "subset", "", "", "subset name for metrics")
}

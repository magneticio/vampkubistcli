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
<<<<<<< HEAD
	"strconv"
=======
	"errors"
>>>>>>> 4483d2651e863f74bf27a4a83987f6548353286f

	"github.com/magneticio/vampkubistcli/logging"
	"github.com/magneticio/vampkubistcli/vampadapter"
	"github.com/spf13/cobra"
)

<<<<<<< HEAD
var PortInt int

// adapterserviceCmd represents the adapterservice command
var adapterserviceCmd = &cobra.Command{
	Use:   "adapterservice",
	Short: "accept logs of the mixer",
	Long: AddAppName(`accept logs of the mixer
Example:
  $AppName adapterservice --port 38355
    `),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logging.Info("Run on port: %v\n", PortInt)
		s, err := vampadapter.NewVampAdapter(strconv.Itoa(PortInt))
=======
// adapterserviceCmd represents the adapterservice command
var adapterserviceCmd = &cobra.Command{
	Use:           "adapterservice",
	Short:         "accept logs of the mixer",
	Long:          `accept logs of the mixer`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Not enough arguments")
		}
		addr := args[0]

		s, err := vampadapter.NewVampAdapter(addr)
>>>>>>> 4483d2651e863f74bf27a4a83987f6548353286f
		if err != nil {
			logging.Error("unable to start server: %v", err)
			return err
		}

		shutdown := make(chan error, 1)
		go func() {
			s.Run(shutdown)
		}()
		_ = <-shutdown
		return nil
	},
}

func init() {
	rootCmd.AddCommand(adapterserviceCmd)
<<<<<<< HEAD
	adapterserviceCmd.Flags().IntVarP(&PortInt, "port", "", 38355, "Port to run the adapter service")

=======
>>>>>>> 4483d2651e863f74bf27a4a83987f6548353286f
}

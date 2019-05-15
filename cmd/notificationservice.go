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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/0xAX/notificator"
	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/magneticio/vampkubistcli/models"
	"github.com/spf13/cobra"
)

var LogoData string = `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`

var desktop bool
var disableOutput bool

// notificationserviceCmd represents the notificationservice command
var notificationserviceCmd = &cobra.Command{
	Use:           "notificationservice",
	Short:         "read notifications from the api",
	Long:          `read notifications from the api`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		iconPath := "/tmp/vamp.png"
		if desktop {
			dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(LogoData))
			buf := new(bytes.Buffer)
			buf.ReadFrom(dec)
			writeFileErr := ioutil.WriteFile(iconPath, buf.Bytes(), 0644)
			if writeFileErr != nil {
				return writeFileErr
			}
		}
		restClient := client.NewRestClient(Config.Url, Config.Token, Config.APIVersion, logging.Verbose, Config.Cert)
		notifications := make(chan models.Notification, 10)
		notify := notificator.New(notificator.Options{
			DefaultIcon: iconPath,
			AppName:     "Vamp Kubist",
		})
		go func() {
			for elem := range notifications {
				if !disableOutput {
					fmt.Println(elem.Text)
				}
				if desktop {
					notify.Push("vamp kubist", elem.Text, iconPath, notificator.UR_NORMAL)
				}
			}
		}()
		err := restClient.ReadNotifications(notifications)
		if err != nil {
			return err
		}
		fmt.Println("End of notifications")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(notificationserviceCmd)

	notificationserviceCmd.Flags().BoolVarP(&desktop, "desktop", "", false, "Generates desktop notificaions")
	notificationserviceCmd.Flags().BoolVarP(&disableOutput, "disable-output", "", false, "Disable printing notificaions to standard output")
}

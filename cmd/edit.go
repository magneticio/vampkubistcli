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
	"os"
	"os/exec"

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/magneticio/vampkubistcli/util"
	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edits a resource",
	Long: AddAppName(`To edit a resource
Run as $AppName edit resourceType ResourceName

Example:
    $AppName edit project myproject
    $AppName edit -p myproject cluster mycluster`),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("Not Enough Arguments")
		}
		Type = args[0]
		Name = args[1]
		restClient := client.NewRestClient(Config.Url, Config.RefreshToken, Config.APIVersion, logging.Verbose, Config.Cert)
		values := make(map[string]string)
		values["project"] = Config.Project
		values["cluster"] = Config.Cluster
		values["virtual_cluster"] = Config.VirtualCluster
		values["application"] = Application
		spec, getSpecError := restClient.GetSpec(Type, Name, SourceFileType, values)
		if getSpecError != nil {
			return getSpecError
		}

		vi := "vim"
		tmpDir := os.TempDir()
		tmpFile, tmpFileErr := ioutil.TempFile(tmpDir, "vampkubist")
		if tmpFileErr != nil {
			return tmpFileErr
		}
		wrriteTempFileError := ioutil.WriteFile(tmpFile.Name(), []byte(spec), 0644)
		if wrriteTempFileError != nil {
			return wrriteTempFileError
		}
		path, err := exec.LookPath(vi)
		if err != nil {
			fmt.Printf("Error %s while looking up for %s!!", path, vi)
		}

		editCommand := exec.Command(path, tmpFile.Name())
		editCommand.Stdin = os.Stdin
		editCommand.Stdout = os.Stdout
		editCommand.Stderr = os.Stderr
		editCommandErr := editCommand.Start()
		if editCommandErr != nil {
			return editCommandErr
		}
		editCommandWaitError := editCommand.Wait()
		if editCommandWaitError != nil {
			return editCommandWaitError
		}

		b, readTempFileError := util.UseSourceUrl(tmpFile.Name())
		if readTempFileError != nil {
			return readTempFileError
		}
		Source := string(b)
		isUpdated, updateError := restClient.Update(Type, Name, Source, SourceFileType, values)
		if !isUpdated {
			return updateError
		}
		fmt.Println(Type + " " + Name + " is edited")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVarP(&SourceFileType, "input", "i", "yaml", "Resource file type yaml or json")
}

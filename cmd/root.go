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
	"fmt"
	"os"

	"github.com/magneticio/vampkubistcli/config"
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Project string
var Cluster string
var VirtualCluster string
var Application string
var Token string
var APIVersion string

var Type string
var Name string
var SourceString string
var SourceFile string
var SourceFileType string
var OutputType string
var Hosts []string

var kubeConfigPath string

// version should be in format d.d.d where d is a decimal number
const Version string = "v0.0.33"

// Backend version is the version this client is tested with
const BackendVersion string = "0.7.12"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   config.AddAppName("$AppName"),
	Short: "A command line client for vamp",
	Long: config.AddAppName(`A command line client for vamp:
  Usage usually follows:
  $AppName create resourceType resourceName -f filepath.yaml
  $AppName update resourceType resourceName -f filepath.yaml
  $AppName delete resourceType resourceName
  eg.:
  $AppName create project myproject -f ./project.yaml
  `),
	SilenceUsage:  true,
	SilenceErrors: true,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	logging.Init(os.Stdout, os.Stderr)

	cobra.OnInitialize(func() { config.Config.InitConfig() })

	if Project != "" {
		config.Config.Project = Project
	}
	if Cluster != "" {
		config.Config.Cluster = Cluster
	}
	if VirtualCluster != "" {
		config.Config.VirtualCluster = VirtualCluster
	}
	if Token != "" {
		config.Config.RefreshToken = Token
	}
	if APIVersion != "" {
		config.Config.APIVersion = APIVersion
	}

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&config.Config.CfgFile, "config", "", config.AddAppName("config file (default is $HOME/.$AppName/config.yaml)"))
	rootCmd.PersistentFlags().StringVarP(&Project, "project", "p", "", "active project")
	rootCmd.PersistentFlags().StringVarP(&Cluster, "cluster", "c", "", "active cluster")
	rootCmd.PersistentFlags().StringVarP(&VirtualCluster, "virtualcluster", "r", "", "active virtual cluster")
	rootCmd.PersistentFlags().StringVarP(&Application, "application", "a", "", "application name for deployments")
	rootCmd.PersistentFlags().StringVarP(&Token, "token", "t", "", "override the login token")
	rootCmd.PersistentFlags().StringVarP(&APIVersion, "api", "", "", "override the api version")
	rootCmd.PersistentFlags().BoolVarP(&logging.Verbose, "verbose", "v", false, "Verbose")

	viper.BindEnv("config", "CONFIG")
}

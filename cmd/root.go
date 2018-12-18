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
	"io/ioutil"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	yaml "gopkg.in/yaml.v2"
)

type config struct {
	Url            string
	Token          string
	Project        string
	Cluster        string
	VirtualCluster string
}

var cfgFile string
var Config config
var Project string
var Cluster string
var VirtualCluster string
var Application string

var Type string
var Name string
var SourceString string
var SourceFile string
var SourceFileType string
var OutputType string
var Debug bool
var Hosts []string

const Version string = "0.0.4"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vamp2cli",
	Short: "A command line client for vamp2",
	Long: `A command line client for vamp2:
  Usage usually follows:
  vamp2cli create resourceType resourceName -f filepath.yaml
  vamp2cli update resourceType resourceName -f filepath.yaml
  vamp2cli delete resourceType resourceName
  eg.:
  vamp2cli create project myproject -f ./project.yaml
  `,
	SilenceUsage: true,
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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vamp2cli.yaml)")
	rootCmd.PersistentFlags().StringVarP(&Project, "project", "p", "", "active project")
	rootCmd.PersistentFlags().StringVarP(&Cluster, "cluster", "c", "", "active cluster")
	rootCmd.PersistentFlags().StringVarP(&VirtualCluster, "virtualcluster", "v", "", "active virtual cluster")
	rootCmd.PersistentFlags().StringVarP(&Application, "application", "a", "", "application name for deployments")

	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "", false, "enable debug on client")
	// rootCmd.PersistentFlags().StringVar(&Server, "server", "default", "server to connect")
	// viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	// Server = viper.Get("server").(string)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func ReadConfigFile() error {
	c := viper.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		fmt.Printf("unable to marshal config to YAML: %v\n", err)
		return err
	}
	err_2 := yaml.Unmarshal(bs, &Config)
	if err_2 != nil {
		fmt.Printf("error: %v\n", err_2)
		return err_2
	}
	if Project != "" {
		Config.Project = Project
	}
	if Cluster != "" {
		Config.Cluster = Cluster
	}
	if VirtualCluster != "" {
		Config.VirtualCluster = VirtualCluster
	}
	// fmt.Printf("Current config: %v \n", Config)
	return nil
}

func WriteConfigFile() error {
	bs, err := yaml.Marshal(Config)
	if err != nil {
		fmt.Printf("unable to marshal config to YAML: %v\n", err)
		return err
	}
	// return string(bs)
	filename := viper.ConfigFileUsed()
	if filename == "" {

		if cfgFile != "" {
			// Use config file from the flag.
			filename = cfgFile
		} else {
			// Find home directory.
			home, err := homedir.Dir()
			if err != nil {
				fmt.Println(err)
				return err
			}
			filename = filepath.FromSlash(home + "/" + ".vamp2cli.yaml")
		}
		// Solves the problem if there is no file viper.ConfigFileUsed() return empty
		os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0644)
	}

	// fmt.Printf("write config to file: %v , config: %v\n", filename, Config)
	err_1 := ioutil.WriteFile(filename, bs, 0644)
	if err_1 != nil {
		return err_1
	}
	return nil
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Search config in home directory with name ".vamp2cli" (without extension).

		viper.AddConfigPath(home)
		viper.SetConfigName(".vamp2cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
		ReadConfigFile()
	}
}

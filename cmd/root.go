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
	"path"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	yaml "gopkg.in/yaml.v2"

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/logging"
	semver "github.com/magneticio/vampkubistcli/semver"
)

type config struct {
	Url            string `yaml:"url,omitempty" json:"url,omitempty"`
	Cert           string `yaml:"cert,omitempty" json:"cert,omitempty"`
	Username       string `yaml:"username,omitempty" json:"username,omitempty"`
	Token          string `yaml:"token,omitempty" json:"token,omitempty"`
	Project        string `yaml:"project,omitempty" json:"project,omitempty"`
	Cluster        string `yaml:"cluster,omitempty" json:"cluster,omitempty"`
	VirtualCluster string `yaml:"virtualcluster,omitempty" json:"virtualcluster,omitempty"`
	APIVersion     string `yaml:"apiversion,omitempty" json:"apiversion,omitempty"`
	TokenStorePath string `yaml:"tokenstorepath,omitempty" json:"tokenstorepath,omitempty"`
}

var cfgFile string
var Config config
var Project string
var Cluster string
var VirtualCluster string
var Application string
var Destination string
var Experiment string
var Port string
var Subset string
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

var TokenStore client.TokenStore

// version should be in format d.d.d where d is a decimal number
const Version string = semver.Version //"v0.0.63"

var AppName string = InitAppName()

// Backend version is the version this client is tested with
const BackendVersion string = "0.8.17"

/*
Application name can change over time so it is made parameteric
*/
func AddAppName(str string) string {
	return strings.Replace(str, "$AppName", AppName, -1)
}

/*
Application name is automacially set to the calling name
*/
func InitAppName() string {
	if len(os.Args) > 0 {
		return FilenameWithoutExtension(os.Args[0])
	}
	return "vamp"
}

// FilenameWithoutExtension extracts the last element without extension from path
func FilenameWithoutExtension(fn string) string {
	fn = path.Base(fn)
	return strings.TrimSuffix(fn, path.Ext(fn))
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   AddAppName("$AppName"),
	Short: "A command line client for vamp",
	Long: AddAppName(`A command line client for vamp:
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

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", AddAppName("config file (default is $HOME/.$AppName/config.yaml)"))
	rootCmd.PersistentFlags().StringVarP(&Project, "project", "p", "", "active project")
	rootCmd.PersistentFlags().StringVarP(&Cluster, "cluster", "c", "", "active cluster")
	rootCmd.PersistentFlags().StringVarP(&VirtualCluster, "virtualcluster", "r", "", "active virtual cluster")
	rootCmd.PersistentFlags().StringVarP(&Application, "application", "a", "", "application name for deployments")
	rootCmd.PersistentFlags().StringVarP(&Token, "token", "t", "", "override the login token")
	rootCmd.PersistentFlags().StringVarP(&APIVersion, "api", "", "", "override the api version")
	rootCmd.PersistentFlags().BoolVarP(&logging.Verbose, "verbose", "v", false, "Verbose")

	viper.BindEnv("config", "CONFIG")

}

func ReadConfig() error {
	c := viper.AllSettings()
	bs, marshalError := yaml.Marshal(c)
	if marshalError != nil {
		return marshalError
	}
	unmarshalError := yaml.Unmarshal(bs, &Config)
	if unmarshalError != nil {
		return unmarshalError
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
	if Token != "" {
		Config.Token = Token
	}
	if APIVersion != "" {
		Config.APIVersion = APIVersion
	}
	return nil
}

func WriteConfigFile() error {
	bs, err := yaml.Marshal(Config)
	if err != nil {
		logging.Error("unable to marshal config to YAML: %v\n", err)
		return err
	}
	filename := viper.ConfigFileUsed()
	if filename == "" {
		if cfgFile != "" {
			// Use config file from the flag.
			filename = cfgFile
		} else {
			// Find home directory.
			home, err := homedir.Dir()
			if err != nil {
				logging.Error("Can not get home dir with error: %v\n", err)
				return err
			}
			path := filepath.FromSlash(home + AddAppName("/.$AppName"))
			if _, err := os.Stat(path); os.IsNotExist(err) {
				os.Mkdir(path, os.ModePerm)
				// If there is a problem here try using MkdirAll
			}
			filename = filepath.FromSlash(path + "/" + "config.yaml")
		}
		// Solves the problem if there is no file viper.ConfigFileUsed() return empty
		os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0644)
	}

	logging.Info("Writing config to config file path: %v\n", filename)
	writeFileError := ioutil.WriteFile(filename, bs, 0644)
	if writeFileError != nil {
		return writeFileError
	}
	return nil
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match
	if cfgFile == "" {
		cfgFile = viper.GetString("config")
	}
	logging.Info("Using Config file path: %v\n", cfgFile)
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, homeDirError := homedir.Dir()
		if homeDirError != nil {
			logging.Error("Can not find home Directory: %v\n", homeDirError)
			os.Exit(1)
		}
		// Search config in home directory with name ".$AppName" (without extension).
		path := filepath.FromSlash(home + AddAppName("/.$AppName"))
		viper.AddConfigPath(path)
		viper.SetConfigName("config")
	}
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logging.Info("Using config file: %v\n", viper.ConfigFileUsed())
	} else {
		logging.Error("Config can not be read due to error: %v\n", err)
	}

	ReadConfig()

	if Config.TokenStorePath == "" {
		tmpfile, tempFileError := ioutil.TempFile("", "tokenstore")
		if tempFileError != nil {
			logging.Error("Token Store file can not be read due to error: %v\n", tempFileError)
		}
		Config.TokenStorePath = tmpfile.Name()
	}
	TokenStore = &client.FileBackedTokenStore{
		Path: Config.TokenStorePath,
	}
}

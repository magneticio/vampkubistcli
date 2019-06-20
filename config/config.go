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

package config

import (
	"github.com/magneticio/vampkubistcli/logging"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type ClientConfig struct {
	Url            string `yaml:"url,omitempty" json:"url,omitempty"`
	Cert           string `yaml:"cert,omitempty" json:"cert,omitempty"`
	Username       string `yaml:"username,omitempty" json:"username,omitempty"`
	RefreshToken   string `yaml:"refresh_token,omitempty" json:"refresh_token,omitempty"`
	AccessToken    string `yaml:"access_token,omitempty" json:"access_token,omitempty"`
	ExpirationTime int64  `yaml:"expiration_time,omitempty" json:"expiration_time,omitempty"`
	Project        string `yaml:"project,omitempty" json:"project,omitempty"`
	Cluster        string `yaml:"cluster,omitempty" json:"cluster,omitempty"`
	VirtualCluster string `yaml:"virtualcluster,omitempty" json:"virtualcluster,omitempty"`
	APIVersion     string `yaml:"apiversion,omitempty" json:"apiversion,omitempty"`
}

var CfgFile string
var Config ClientConfig

var AppName string = InitAppName()

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
		if CfgFile != "" {
			// Use config file from the flag.
			filename = CfgFile
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
func InitConfig() {
	viper.AutomaticEnv() // read in environment variables that match
	if CfgFile == "" {
		CfgFile = viper.GetString("config")
	}
	logging.Info("Using Config file path: %v\n", CfgFile)
	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
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
}

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
		return os.Args[0]
	}
	return "vamp"
}

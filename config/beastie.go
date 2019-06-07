// Copyright © 2018 Chris Custine <ccustine@apache.org>
//
// Licensed under the Apache License, version 2.0 (the "License");
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
	"fmt"
	"github.com/kellydunn/golang-geo"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

var (
	ConfigFile string
)

type BeastInfo struct {
	Sources   []Source
	Homepos   *geo.Point
	Latitude  float64  `yaml:"baseLat"`
	Longitude float64  `yaml:"baseLon"`
	Debug     bool     `yaml:"debug"`
	Metrics   bool     `yaml:"metrics"`
	Outputs   []string `yaml:"output"`
	RtlInput  bool     `yaml:"rtl"`
}

type Source struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

const (
	DEBUG      = "debug"
	METRICS    = "metrics"
	BEAST_HOST = "adsbHost"
	BEAST_PORT = "adsbPort"
	MLAT_HOST  = "mlatHost"
	MLAT_PORT  = "mlatPort"
	BASELAT    = "lat"
	BASELON    = "lon"
	CONFIGFILE = "config"
	OUTPUT     = "out"
)

func LoadConfig() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if ConfigFile != "" {
		// Use viper file from the flag.
		viper.SetConfigFile(ConfigFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath(home + "/.beastie")
	}

	if err := viper.ReadInConfig(); err != nil {
		// Maybe for testing
		//if err := viper.ReadConfig(bytes.NewBuffer(yamlExample)); err != nil {
		log.Warnf("Can't read viper: %s\n", err)
	}
}

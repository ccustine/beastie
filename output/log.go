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

package output

import (
	"encoding/json"
	"github.com/ccustine/beastie/config"
	"github.com/ccustine/beastie/types"
	"github.com/kellydunn/golang-geo"
	log "github.com/sirupsen/logrus"
	"os"
	"sort"
)

var (
	aclog *log.Logger
)

const (
	LOG = "log"
)

type LogOutput struct {
	ACLogFile *os.File
	Beastinfo *config.BeastInfo
	Aclog *log.Logger
}

func NewLogOutput(info *config.BeastInfo) *LogOutput {
	here = geo.NewPoint(info.Latitude, info.Longitude)

	file, err := os.OpenFile("/tmp/aclog.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		log.Fatal(err)
	}
	//defer file.Close()
	aclog = log.New()
	aclog.SetOutput(file)

	return &LogOutput{Beastinfo:info, Aclog:aclog, ACLogFile:file}
}

func (o LogOutput) UpdateDisplay(knownAircraft []*types.AircraftData) {
	//var b strings.Builder

	sortedAircraft := make(AircraftList, 0, len(knownAircraft))

	for _, aircraft := range knownAircraft { //.Copy() {
		sortedAircraft = append(sortedAircraft, aircraft)
	}

	sort.Sort(sortedAircraft)

	for _, aircraft := range sortedAircraft {
		jsonString, _ := json.MarshalIndent(aircraft, "", "\t")
		if _, err := o.ACLogFile.Write(jsonString); err != nil {
			log.Warnf("Unable to write to json file: %s", err)
		}
	}
}

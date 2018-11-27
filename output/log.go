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

func (o LogOutput) UpdateDisplay(knownAircraft types.AircraftMap) {
	//var b strings.Builder

	sortedAircraft := make(AircraftList, 0, knownAircraft.Len())

	knownAircraft.RLocker().Lock()
	for _, aircraft := range knownAircraft.Range() {
		sortedAircraft = append(sortedAircraft, aircraft)
	}
	knownAircraft.RLocker().Unlock()

	sort.Sort(sortedAircraft)

	for _, aircraft := range sortedAircraft {
/*		evict := time.Since(aircraft.LastPing) > (time.Duration(59) * time.Second)

		if evict {
			if o.Beastinfo.Debug {
				log.Debugf("Evicting %d", aircraft.IcaoAddr)
			}
			knownAircraft.Delete(aircraft.IcaoAddr)
			continue
		}
*/
		/*		aircraftHasLocation := aircraft.Latitude != math.MaxFloat64 &&
			aircraft.Longitude != math.MaxFloat64
		aircraftHasAltitude := aircraft.Altitude != math.MaxInt32
*/

		//displayTable.SetStyle(simpletable.StyleCompact)
		//b.WriteString(spew.Sprintln(aircraft))
		//b.WriteString("\n")
		jsonString, _ := json.MarshalIndent(aircraft, "", "\t")
		o.ACLogFile.WriteString(string(jsonString))
		//o.Aclog.Info(string(jsonString))
	}
}

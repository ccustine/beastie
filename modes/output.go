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

package modes

import (
	"fmt"
	"github.com/InVisionApp/tabular"
	"github.com/kellydunn/golang-geo"
	. "github.com/logrusorgru/aurora"
	"github.com/sirupsen/logrus"
	"io"
	"math"
	"sort"
	"strings"
	"time"
)

var (
	tab     tabular.Table
	table	tabular.Output
)

func init() {
	tab = tabular.New()
	tab.Col("row", "#", 3)
	tab.Col("icao", "ICAO", 6)
	tab.Col("call", "Callsign", 8)
	tab.Col("latlon", "Lat/Lon", 17)
	//tab.Col("mlat", "MLAT", 4)
	tab.ColRJ("alt", "Altitude", 8)
	tab.ColRJ("roc", "RoC", 8)
	tab.ColRJ("spd", "Speed", 5)
	tab.ColRJ("hdg", "Hdg", 3)
	tab.ColRJ("dst", "Distance", 8)
	tab.ColRJ("last", "Last Seen", 9)
	tab.ColRJ("rssi", "RSSI", 7)

	table = tab.Parse("row", "icao", "call", "latlon",  /*"mlat",*/   "alt", "roc", "spd", "hdg", "dst", "last", "rssi")
}

func durationSecondsElapsed(since time.Duration) string {
	sec := uint8(since.Seconds())
	if sec == math.MaxUint8 {
		return "-"
	} else {
		return fmt.Sprintf("%4d", sec)
	}
}

func updateDisplay(knownAircraft *AircraftMap, writer io.Writer) {
	var b strings.Builder

	b.WriteString(Bold(Cyan(table.Header)).String())
	b.WriteString("\n")
	b.WriteString(Bold(Cyan(table.SubHeader)).String())
	b.WriteString("\n")

	sortedAircraft := make(aircraftList, 0, knownAircraft.Len())

	knownAircraft.RLocker().Lock()
	for _, aircraft := range knownAircraft.Range() {
		sortedAircraft = append(sortedAircraft, aircraft)
	}
	knownAircraft.RLocker().Unlock()

	sort.Sort(sortedAircraft)

	for i, aircraft := range sortedAircraft {
		old := time.Since(aircraft.lastPos) > (time.Duration(10) * time.Second)
		doomed := time.Since(aircraft.lastPos) > (time.Duration(20) * time.Second)
		evict := time.Since(aircraft.lastPos) > (time.Duration(59) * time.Second)

		if evict {
			if info.Debug {
				logrus.Debugf("Evicting %d", aircraft.icaoAddr)
			}
			knownAircraft.Delete(aircraft.icaoAddr)
			continue
		}

		aircraftHasLocation := aircraft.latitude != math.MaxFloat64 &&
					aircraft.longitude != math.MaxFloat64
		aircraftHasAltitude := aircraft.altitude != math.MaxInt32

		if aircraft.callsign != "" || aircraftHasLocation || aircraftHasAltitude {
			var sLatLon string
			var sAlt string

			isMlat := ""
			if aircraft.mlat {
				isMlat = "*"
			}

			if aircraftHasLocation {
				sLatLon = fmt.Sprintf("%3.3f, %3.3f%s", aircraft.latitude, aircraft.longitude, isMlat)
			} else {
				sLatLon = "---.------,---.------"
			}
			if aircraftHasAltitude {
/*				var altUnit string
				switch aircraft.altUnit {
				case 0:
					altUnit = "ft"
				case 1:
					altUnit = "m"
				}
*/

				// TODO: This is noisy, need to figure out how to smooth and watch trending
				var vrs string
				switch aircraft.vertRateSign {
				case 0:
					vrs = "➚"
				case 1:
					vrs = "➘"
				default:
					vrs = ""
				}

				sAlt = fmt.Sprintf("%d %s", aircraft.altitude, vrs)
				//sAlt = fmt.Sprintf("%d%s %s", aircraft.altitude, altUnit, vrs)
			} else {
				sAlt = "-----"
			}

			acpos := geo.NewPoint(aircraft.latitude, aircraft.longitude)
			homepos := geo.NewPoint(info.Latitude, info.Longitude)
			dist := homepos.GreatCircleDistance(acpos)

			distance := dist * 0.539957 // nm //0.621371 - statue mile

			//tPing := time.Since(aircraft.lastPing)
			tPos := time.Since(aircraft.lastPos)

			if !old && !doomed {
				b.WriteString(Cyan(
					fmt.Sprintf(table.Format, i+1, fmt.Sprintf("%06x", aircraft.icaoAddr), aircraft.callsign,
						sLatLon, /*isMlat,*/ sAlt, aircraft.vertRate, aircraft.speed, //fmt.Sprintf("%5.0f", aircraft.speed),
						aircraft.heading, //fmt.Sprintf("%3.0f", aircraft.heading),
						fmt.Sprintf("%3.1f", distance),
						fmt.Sprintf("%2d", uint8(tPos.Seconds())), fmt.Sprintf("%.1f", aircraft.rssi))).String(),
						)
			} else if old && !doomed {
				b.WriteString(Brown(
					fmt.Sprintf(table.Format, i+1, fmt.Sprintf("%06x", aircraft.icaoAddr), aircraft.callsign,
						sLatLon, /*isMlat, */ sAlt, aircraft.vertRate, aircraft.speed, //fmt.Sprintf("%5.0f", aircraft.speed),
						aircraft.heading, //fmt.Sprintf("%3.0f", aircraft.heading),
						fmt.Sprintf("%3.1f", distance),
						fmt.Sprintf("%2d", uint8(tPos.Seconds())), fmt.Sprintf("%.1f", aircraft.rssi))).String())
			} else if doomed {
				b.WriteString(Red(
					fmt.Sprintf(table.Format, i+1, fmt.Sprintf("%06x", aircraft.icaoAddr), aircraft.callsign,
						sLatLon, /*isMlat,*/ sAlt, aircraft.vertRate, aircraft.speed, //fmt.Sprintf("%5.0f", aircraft.speed),
						aircraft.heading, //fmt.Sprintf("%3.0f", aircraft.heading),
						fmt.Sprintf("%3.1f", distance),
						fmt.Sprintf("%2d", uint8(tPos.Seconds())), fmt.Sprintf("%.1f", aircraft.rssi))).String())
			}

		}
	}

	b.WriteString(fmt.Sprintf("Message Rate (Good) 1 Min: %.1f/s\n", GoodRate.Rate1()))
	b.WriteString(fmt.Sprintf("Message Rate (Bad)  1 Min: %.1f/s\n", BadRate.Rate1()))
	b.WriteString(fmt.Sprintf("Message Count - ModeA/C: %d\n", ModeACCnt.Count()))
	b.WriteString(fmt.Sprintf("Message Count - ModeS Short: %d\n", ModesShortCnt.Count()))
	b.WriteString(fmt.Sprintf("Message Count - ModeS Long: %d\n", ModesLongCnt.Count()))

	writer.Write([]byte(b.String()))
}

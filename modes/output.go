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
	"github.com/alexeyco/simpletable"
	"github.com/aybabtme/rgbterm"
	"github.com/ccustine/uilive"
	"github.com/kellydunn/golang-geo"
	//. "github.com/logorgru/aurora"
	log "github.com/sirupsen/logrus"
	"image/color"
	"math"
	"sort"
	"strings"
	"time"
)

var (
	displayTable *simpletable.Table
)

func init() {
	displayTable = simpletable.New()

	displayTable.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "#"},
			{Align: simpletable.AlignCenter, Text: "ICAO"},
			{Align: simpletable.AlignLeft, Text: "Call"},
			{Align: simpletable.AlignLeft, Text: "Squawk"},
			{Align: simpletable.AlignCenter, Text: "Lat/Lon"},
			{Align: simpletable.AlignRight, Text: "Alt"},
			{Align: simpletable.AlignRight, Text: "Rate"},
			{Align: simpletable.AlignRight, Text: "Speed"},
			{Align: simpletable.AlignCenter, Text: "Hdg"},
			{Align: simpletable.AlignRight, Text: "Dist"},
			{Align: simpletable.AlignRight, Text: "Last"},
			//{Align: simpletable.AlignRight, Text: "RSSI"},
		},
	}
}

var colorPalette = map[string][]color.RGBA{
	"orange": {{255, 175, 0, 255}, {215, 135, 0, 255}, {175, 95, 0, 255}},
	"white": {{255, 255, 255, 255},{175, 175, 175, 255}, {95, 95, 95, 255}},
	"red": {{255, 0, 0, 255},{175, 0,0,255},{95,0, 0,255}},
	"rainbow": {{255, 255, 255, 255},{255, 170, 30, 255},{255, 100,100,255}},
}

func updateDisplay(knownAircraft *AircraftMap, writer *uilive.Writer) {
	displayTable.Body.Cells = [][]*simpletable.Cell{}
	var b strings.Builder

	sortedAircraft := make(aircraftList, 0, knownAircraft.Len())

	knownAircraft.RLocker().Lock()
	for _, aircraft := range knownAircraft.Range() {
		sortedAircraft = append(sortedAircraft, aircraft)
	}
	knownAircraft.RLocker().Unlock()

	sort.Sort(sortedAircraft)

	for i, aircraft := range sortedAircraft {
		evict := time.Since(aircraft.lastPing) > (time.Duration(59) * time.Second)

		if evict {
			if info.Debug {
				log.Debugf("Evicting %d", aircraft.icaoAddr)
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
				// TODO: This is noisy, need to figure out how to smooth and watch trending
				var vrs string
				if aircraft.vertRate >= 250 {
					switch aircraft.vertRateSign {
					case 0:
						vrs = "➚"
					case 1:
						vrs = "➘"
					default:
						vrs = ""
					}
				} else {
					vrs = ""
				}

				sAlt = fmt.Sprintf("%d %s", aircraft.altitude, vrs)
			} else {
				sAlt = "-----"
			}

			acpos := geo.NewPoint(aircraft.latitude, aircraft.longitude)
			homepos := geo.NewPoint(info.Latitude, info.Longitude)
			dist := homepos.GreatCircleDistance(acpos)

			distance := dist * 0.539957 // nm //0.621371 - statue mile

			theme := colorPalette["rainbow"]

			tPing := time.Since(aircraft.lastPing)
			var rowcolor color.Color
			stale := tPing > (time.Duration(10) * time.Second)
			pendingEvict := tPing > (time.Duration(35) * time.Second)
			if !stale && !pendingEvict {
				rowcolor = theme[0]
			} else if stale && !pendingEvict {
				rowcolor = theme[1]
			} else if pendingEvict {
				rowcolor = theme[2]
			}

			tPos := time.Since(aircraft.lastPos)
			var poscolor color.Color
			posstale := tPos > (time.Duration(10) * time.Second)
			pospendingEvict := tPos > (time.Duration(20) * time.Second)
			//posevict := time.Since(aircraft.lastPos) > (time.Duration(59) * time.Second)

			if !posstale && !pospendingEvict {
				poscolor = theme[0]
			} else if posstale && !pospendingEvict {
				poscolor = theme[1]
			} else if pospendingEvict {
				poscolor = theme[2]
			}

			var vertRate string
			if aircraft.vertRate >= 250 {
				vertRate = fmt.Sprintf("%d", aircraft.vertRate)
			} else {
				vertRate = ""
			}

			var squawk string
			if aircraft.squawk > 0 {
				squawk = fmt.Sprintf("%04x", aircraft.squawk)
			} else {
				squawk = ""
			}

			r := []*simpletable.Cell{
				{Align: simpletable.AlignRight, Text: colorize(fmt.Sprintf("%d", i+1), rowcolor)},
				{Text: colorize(fmt.Sprintf("%06x", aircraft.icaoAddr), rowcolor)},
				{Text: colorize(aircraft.callsign, rowcolor)},
				//{Text: colorize(fmt.Sprintf("%x", i), rowcolor)},
				{Text: colorize(squawk, rowcolor)},
				{Text: colorize(sLatLon, poscolor)},
				{Text: colorize(sAlt, rowcolor)},
				{Text: colorize(vertRate, rowcolor)},
				{Text: colorize(fmt.Sprintf("%d", aircraft.speed), rowcolor)},
				{Text: colorize(fmt.Sprintf("%d", aircraft.heading), rowcolor)},
				{Text: colorize(fmt.Sprintf("%3.1f", distance), rowcolor)},
				{Text: colorize(fmt.Sprintf("%2d", uint8(tPing.Seconds())), rowcolor)},
				//{Text: colorize(fmt.Sprintf("%.1f", aircraft.rssi), rowcolor)},
			}

			displayTable.Body.Cells = append(displayTable.Body.Cells, r)

		}
	}

	displayTable.Footer = &simpletable.Footer{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Span: 4, Text: fmt.Sprintf("Message Rate (Good): %.1f/s\nMessage Rate (Bad) : %.1f/s", GoodRate.Rate1(), BadRate.Rate1())},
			{Align: simpletable.AlignLeft, Span: 2, Text: fmt.Sprintf("")},
			{Align: simpletable.AlignLeft, Span: 5, Text: fmt.Sprintf("Message Count - Mode A/C:    %d\nMessage Count - ModeS Short: %d\nMessage Count - ModeS Long:  %d", ModeACCnt.Count(), ModesShortCnt.Count(), ModesLongCnt.Count())},
		},
	}

	displayTable.SetStyle(simpletable.StyleCompact)
	b.WriteString(displayTable.String())
	b.WriteString("\n")
	writer.Write([]byte(b.String()))
}

func colorize(text string, newColor color.Color) string {
	r, g, b, _ :=newColor.RGBA()
	return rgbterm.FgString(text, uint8(r), uint8(g), uint8(b))
}
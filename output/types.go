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
	"fmt"
	"github.com/ccustine/beastie/types"
	"github.com/kellydunn/golang-geo"
	"math"
)

var (
	here *geo.Point
)

type AircraftList []*types.AircraftData

type Output interface {
	UpdateDisplay(aircraftMap *types.AircraftMap)
	//NewTableOutput(*config.BeastInfo) *Output
}

// List utilities
func (a AircraftList) Len() int {
	return len(a)
}
func (a AircraftList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a AircraftList) Less(i, j int) bool {
	/*	if *sortMode == sortModeLastPos {
			// t1 later than t2 means that t1 is more recent
			return a[i].lastPos.After(a[j].lastPos)

		} else if *sortMode == sortModeDistance {
	*/
	if a[i].Latitude != math.MaxFloat64 && a[j].Latitude != math.MaxFloat64 {
		return sortAircraftByDistance(a, i, j)
	} else if a[i].Latitude != math.MaxFloat64 && a[j].Latitude == math.MaxFloat64 {
		return true
	} else if a[i].Latitude == math.MaxFloat64 && a[j].Latitude != math.MaxFloat64 {
		return false
	}
	/*		return sortAircraftByCallsign(a, i, j)
		} else if *sortMode == sortModeCallsign {
			return sortAircraftByCallsign(a, i, j)
		}
	*/	return false
}

func sortAircraftByDistance(a AircraftList, i, j int) bool {
	pi := geo.NewPoint(a[i].Latitude, a[i].Longitude)
	pj := geo.NewPoint(a[j].Latitude, a[j].Longitude)
	return here.GreatCircleDistance(pi) < here.GreatCircleDistance(pj)
	// Need to compare speeds of these two methods for GreatCircle
	/*	return greatcircle(a[i].latitude, a[i].longitude, info.Latitude, info.Longitude) <
			greatcircle(a[j].latitude, a[j].longitude, info.Latitude, info.Longitude)
	*/
}
func sortAircraftByCallsign(a AircraftList, i, j int) bool {
	if a[i].Callsign != "" && a[j].Callsign != "" {
		return a[i].Callsign < a[j].Callsign
	} else if a[i].Callsign != "" && a[j].Callsign == "" {
		return true
	} else if a[i].Callsign == "" && a[j].Callsign != "" {
		return false
	}
	return fmt.Sprintf("%06x", a[i].IcaoAddr) < fmt.Sprintf("%06x", a[j].IcaoAddr)
}

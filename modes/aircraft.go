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
	"github.com/kellydunn/golang-geo"
	"math"
	"sync"
	"time"
)

type aircraftData struct {
	icaoAddr uint32

	callsign string
	squawk   uint32

	eRawLat uint32
	eRawLon uint32
	oRawLat uint32
	oRawLon uint32

	latitude  float64
	longitude float64
	altitude  int32
	altUnit   uint

	//ewd            uint  // 0 = East, 1 = West.
	//ewv            int32 // E/W velocity.
	//nsd            uint  // 0 = North, 1 = South.
	//nsv            int32 // N/S velocity.
	vertRateSource uint  // Vertical rate source.
	vertRateSign   uint  // Vertical rate sign.
	vertRate       int32 // Vertical rate.
	speed          int32
	heading        int32
	headingIsValid bool

	lastPing time.Time
	lastPos  time.Time

	rssi float64

	mlat bool
}
type aircraftList []*aircraftData
//type AircraftMap map[uint32]*aircraftData

type AircraftMap struct {
	sync.RWMutex
	internal map[uint32]*aircraftData
}

func NewAircraftMap() *AircraftMap {
	return &AircraftMap{
		internal: make(map[uint32]*aircraftData),
	}
}

func (am *AircraftMap) Load(key uint32) (value *aircraftData, ok bool) {
	am.RLock()
	result, ok := am.internal[key]
	am.RUnlock()
	return result, ok
}

func (am *AircraftMap) Delete(key uint32) {
	am.Lock()
	delete(am.internal, key)
	am.Unlock()
}

func (am *AircraftMap) Store(key uint32, value *aircraftData) {
	am.Lock()
	am.internal[key] = value
	am.Unlock()
}

func (am *AircraftMap) Len() (length int) {
	am.RLock()
	result := len(am.internal)
	am.RUnlock()
	return result
}

func (am *AircraftMap) Range() (map[uint32]*aircraftData) {
	return am.internal
}


func (a aircraftList) Len() int {
	return len(a)
}
func (a aircraftList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a aircraftList) Less(i, j int) bool {
/*	if *sortMode == sortModeLastPos {
		// t1 later than t2 means that t1 is more recent
		return a[i].lastPos.After(a[j].lastPos)

	} else if *sortMode == sortModeDistance {
*/		if a[i].latitude != math.MaxFloat64 && a[j].latitude != math.MaxFloat64 {
			return sortAircraftByDistance(a, i, j)
		} else if a[i].latitude != math.MaxFloat64 && a[j].latitude == math.MaxFloat64 {
			return true
		} else if a[i].latitude == math.MaxFloat64 && a[j].latitude != math.MaxFloat64 {
			return false
		}
/*		return sortAircraftByCallsign(a, i, j)
	} else if *sortMode == sortModeCallsign {
		return sortAircraftByCallsign(a, i, j)
	}
*/	return false
}

func sortAircraftByDistance(a aircraftList, i, j int) bool {
	p := geo.NewPoint(info.Latitude, info.Longitude)
	pi := geo.NewPoint(a[i].latitude, a[i].longitude)
	pj := geo.NewPoint(a[j].latitude, a[j].longitude)
	return p.GreatCircleDistance(pi) < p.GreatCircleDistance(pj)
	// Need to compare speeds of these two methods for GreatCircle
/*	return greatcircle(a[i].latitude, a[i].longitude, info.Latitude, info.Longitude) <
		greatcircle(a[j].latitude, a[j].longitude, info.Latitude, info.Longitude)
*/
}
func sortAircraftByCallsign(a aircraftList, i, j int) bool {
	if a[i].callsign != "" && a[j].callsign != "" {
		return a[i].callsign < a[j].callsign
	} else if a[i].callsign != "" && a[j].callsign == "" {
		return true
	} else if a[i].callsign == "" && a[j].callsign != "" {
		return false
	}
	return fmt.Sprintf("%06x", a[i].icaoAddr) < fmt.Sprintf("%06x", a[j].icaoAddr)
}

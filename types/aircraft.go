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

package types

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"
)

type AircraftData struct {
	IcaoAddr uint32

	Callsign string
	Squawk   uint32

	ERawLat uint32
	ERawLon uint32
	ORawLat uint32
	ORawLon uint32

	Latitude  float64
	Longitude float64
	Altitude  int32
	AltUnit   uint

	//ewd            uint  // 0 = East, 1 = West.
	//ewv            int32 // E/W velocity.
	//nsd            uint  // 0 = North, 1 = South.
	//nsv            int32 // N/S velocity.
	VertRateSource uint  // Vertical rate source.
	VertRateSign   uint  // Vertical rate sign.
	VertRate       int32 // Vertical rate.
	Speed          int32
	Heading        int32
	HeadingIsValid bool

	LastPing time.Time
	LastPos  time.Time

	Rssi float64

	Mlat    bool
	IsValid bool
}
//type AircraftMap map[uint32]*AircraftData

func (a *AircraftData) MarshalJSON() ([]byte, error) {
	type Alias AircraftData

	var vertRate string
	if a.VertRate >= 250 {
		vertRate = fmt.Sprintf("%d", a.VertRate)
	} else {
		vertRate = ""
	}

	var squawk string
	if a.Squawk > 0 {
		squawk = fmt.Sprintf("%04x", a.Squawk)
	} else {
		squawk = ""
	}

	var sLat, sLong string
	if a.Latitude != math.MaxFloat64 &&
		a.Longitude != math.MaxFloat64 {
		sLat = fmt.Sprintf("%3.3f", a.Latitude)
		sLong = fmt.Sprintf("%3.3f", a.Longitude)
	} else {
		sLat = ""
		sLong = ""
	}

	return json.Marshal(&struct {
		IcaoAddr     string `json:"icao"`
		Squawk       string `json:"xpdr,omitempty"`
		VertRate     string `json:"vrt,omitempty"`
		Latitude     string `json:"lat,omitempty"`
		Longitude    string `json:"lon,omitempty"`
		MLat         bool   `json:"mlat,omitempty"`
		Altitude     int32  `json:"alt,omitempty"`
		VertRateSign uint   `json:"vrtsgn,omitempty"`
		Speed        int32  `json:"spd,omitempty"`
		Heading      int32  `json:"hdg,omitempty"`
		Distance      int32  `json:"rng,omitempty"`
		Callsign      string  `json:"call,omitempty"`
		//*Alias
	}{
		IcaoAddr:     fmt.Sprintf("%06x", a.IcaoAddr),
		Squawk:       squawk,
		VertRate:     vertRate,
		Latitude:     sLat,
		Longitude:    sLong,
		MLat:         a.Mlat,
		Altitude:     a.Altitude,
		VertRateSign: a.VertRateSign,
		Speed:        a.Speed,
		Heading:      a.Heading,
		Callsign:      a.Callsign,
		//Alias:    (*Alias)(a),
	})
}

type AircraftMap struct {
	sync.RWMutex
	internal map[uint32]*AircraftData
}

func NewAircraftMap() *AircraftMap {
	return &AircraftMap{
		internal: make(map[uint32]*AircraftData),
	}
}

func (am *AircraftMap) Load(key uint32) (value *AircraftData, ok bool) {
	am.RLock()
	//defer am.RUnlock()
	result, ok := am.internal[key]
	am.RUnlock()
	return result, ok
}

func (am *AircraftMap) Delete(key uint32) {
	am.Lock()
	//defer am.Unlock()
	delete(am.internal, key)
	am.Unlock()
}

func (am *AircraftMap) Store(key uint32, value *AircraftData) {
	am.Lock()
	//defer am.Unlock()
	am.internal[key] = value
	am.Unlock()
}

func (am *AircraftMap) Len() (length int) {
	am.RLock()
	//defer am.RUnlock()
	result := len(am.internal)
	am.RUnlock()
	return result
}

func (am *AircraftMap) Copy() ([]*AircraftData) {
	am.RLock()
	//defer am.RUnlock()
	result := make([]*AircraftData, 0, len(am.internal))
	for _, ac := range am.internal {
		result = append(result, ac)
	}
	am.RUnlock()
	return result
}
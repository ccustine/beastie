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
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	"math"
	"sync"
	"time"
)

const (
	TILE38 = "tile38"
)

type Tile38Output struct {
	aircraftList []*types.AircraftData
	lock         sync.RWMutex
	rc           redis.Conn
}

func NewTile38Output() *Tile38Output {
	rc, err := redis.Dial("tcp", "127.0.0.1:9851", redis.DialReadTimeout(1 * time.Second), redis.DialWriteTimeout(1 * time.Second))
	if err != nil {
		logrus.Warnf("error: tile38 - ", err)
		return nil
	}

	tile38Output := &Tile38Output{rc: rc}

	//defer rc.Close()

	return tile38Output
}

func (o *Tile38Output) UpdateDisplay(knownAircraft []*types.AircraftData) { //*types.AircraftMap) {
	//logrus.Warn("Step 1")

	connErr := o.rc.Err()
	if connErr != nil {
		logrus.Errorf("Tile38 connection error: %s", connErr.Error())
		err := o.rc.Close()
		if err != nil {
			logrus.Warnf("error on Close: tile38 - ", err)
			return
		}


		rc, err := redis.Dial("tcp", "127.0.0.1:9851", redis.DialReadTimeout(1 * time.Second), redis.DialWriteTimeout(1 * time.Second))
		if err != nil {
			logrus.Warnf("error on Dial: tile38 - ", err)
			return
		}
		o.rc = rc
	}

	//logrus.Warn("Step 1a")

	//sortedAircraft := make(AircraftList, 0, aircraftList.Len())
	//aircraftList := make([]AircraftData, 0, len(knownAircraft)) //.Len())

	//for _, aircraft := range knownAircraft { //.Copy() {
	//	aircraftList = append(aircraftList, *aircraft)
	//}

	//o.lock.Lock()
	o.aircraftList = knownAircraft //.Copy()
	//o.lock.Unlock()

	//logrus.Warn("Step 2")

	connErr = o.rc.Err()
	if connErr != nil {
		logrus.Errorf("Tile38 connection error: %s", connErr.Error())
		return
	}

	//logrus.Warn("Step 3")

	process := false
	//o.lock.RLock()
	for i, aircraft := range o.aircraftList {
		logrus.Warnf("Processing AC %d", i)

		aircraftHasLocation := aircraft.Latitude != math.MaxFloat64 &&
			aircraft.Longitude != math.MaxFloat64
		// This hides ac with no pos from the display
		if !aircraftHasLocation {
			continue
		}

		process = true
		//logrus.Warnf("Processing AC %d, has pos", i)

		err := o.rc.Send("SET", "aircraft", fmt.Sprintf("%06x", aircraft.IcaoAddr),
			"EX", 59,
			"FIELD", "spd", aircraft.Speed,
			"FIELD", "hdg", aircraft.Heading,
			"POINT", aircraft.Latitude, aircraft.Longitude, aircraft.Altitude,
		)
		//logrus.Warn("Step 3a")
		if err != nil {
			logrus.Warnf("error: tile38 Set CMD - ", err)
		}

		/*		err = o.rc.Send("SET", "aircraft", fmt.Sprintf("%06x", aircraft.icao) + ":call",
					"STRING", aircraft.Callsign,
				)
				if err != nil {
					logrus.Warnf("error: tile38 Set CMD(Call) - ", err)
				}
		*/
	}
	//o.lock.RUnlock()

	//logrus.Warn("Step 3b (Should happen once)")

	// All of the Redis API calls below only need to happen if we actually sent anything above
	if !process {
		return
	}

	//logrus.Warn("Step 4")

	err := o.rc.Flush()
	//logrus.Warn("Step 4a")
	if err != nil {
		logrus.Warnf("error: tile38 flush on Set CMD(Call) - ", err)
	}

	//logrus.Warn("Step 5")

	reply, err := redis.String(o.rc.Receive())
	if err != nil {
		logrus.Errorf("Err: %s Reply: %s", err, reply)
	}

	//logrus.Warn("Step 6")

	//logrus.Infof("RESP Reply %s", reply)
}

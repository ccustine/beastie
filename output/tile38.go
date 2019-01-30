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
	"sync"
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
	rc, err := redis.Dial("tcp", "127.0.0.1:9851")
	if err != nil {
		logrus.Warnf("error: tile38 - ", err)
		return nil
	}

	tile38Output := &Tile38Output{rc: rc}

	//defer rc.Close()

	return tile38Output
}

func (o Tile38Output) UpdateDisplay(knownAircraft []*types.AircraftData) { //*types.AircraftMap) {
	//sortedAircraft := make(AircraftList, 0, aircraftList.Len())

	//o.lock.Lock()
	o.aircraftList = knownAircraft //.Copy()
	//o.lock.Unlock()

	for _, aircraft := range o.aircraftList {
		err := o.rc.Send("SET", "aircraft", fmt.Sprintf("%06x", aircraft.IcaoAddr),
			"EX", 59,
			"FIELD", "spd", aircraft.Speed,
			"FIELD", "hdg", aircraft.Heading,
			"POINT", aircraft.Latitude, aircraft.Longitude, aircraft.Altitude,
		)
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

	err := o.rc.Flush()
	if err != nil {
		logrus.Warnf("error: tile38 Set CMD(Call) - ", err)
	}

	reply, err := o.rc.Receive()
	if err != nil {
		logrus.Warnf("error: tile38 Set CMD(Call) - ", err)
		logrus.Warnf("error: tile38 Resp - ", reply)
	}

	//logrus.Infof("RESP Reply %s", reply)
}

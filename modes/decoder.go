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
	"encoding/binary"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"

	//"fmt"
	"math"
	"time"
)

func decodeModeS(message []byte, isMlat bool, sig float64) aircraftData { //}, knownAircraft *AircraftMap) {
	df := uint((message[0] & 0xF8) >> 3)

	var aircraft aircraftData
	var aircraftExists bool
	aircraft.vertRateSign = math.MaxUint32
	icaoAddr := uint32(math.MaxUint32)
	//altCode := uint16(math.MaxUint16)
	//altitude := int32(math.MaxInt32)

	var msgType string

	metrics.GetOrRegisterCounter(fmt.Sprintf("DF %02d", df), nil).Inc(1)

	switch df {
	case 0:
	 msgType = "short air-air surveillance (TCAS)"
	case 4:
	 msgType = "surveillance, altitude reply"
	case 5:
	 msgType = "surveillance, Mode A identity reply"
	case 11:
	msgType = "All-Call reply containing aircraft address"
	case 16:
	 msgType = "long air-air surveillance (TCAS)"
	case 17:
	 msgType = "extended squitter"
	case 18:
	 msgType = "TIS-B"
	case 19:
	 msgType = "military extended squitter"
	case 20:
	 msgType = "Comm-B including altitude reply"
	case 21:
	 msgType = "Comm-B reply including Mode A identity"
	case 22:
	 msgType = "military use"
	case 24:
	 msgType = "special long msg"
	default:
	 msgType = "unknown"
	}
	if info.Debug {
		fmt.Printf("UF: %d\n", df)
		fmt.Printf("UF: %08s\n", strconv.FormatInt(int64(df), 2))
		fmt.Println(msgType)
	}

	//if df == 0 || df == 4 || df == 11 || df == 17 || df == 18 {
	if df == 11 || df == 17 || df == 18 {
		icaoAddr = uint32(message[1]) << 16 | uint32(message[2]) << 8 | uint32(message[3]) //New
		if info.Debug {
			fmt.Printf("ICAO: %06x\n", icaoAddr)
		}
	}

	if icaoAddr != math.MaxUint32 {
		var ptrAircraft *aircraftData
		ptrAircraft, aircraftExists = knownAircraft.Load(icaoAddr)
		if !aircraftExists {
			// initialize some values
			aircraft = aircraftData{
				icaoAddr:  icaoAddr,
				oRawLat:   math.MaxUint32,
				oRawLon:   math.MaxUint32,
				eRawLat:   math.MaxUint32,
				eRawLon:   math.MaxUint32,
				latitude:  math.MaxFloat64,
				longitude: math.MaxFloat64,
				altitude:  math.MaxInt32,
				callsign:  "",
				mlat:      isMlat,
				rssi: sig,
			}
		} else {
			aircraft = *ptrAircraft
			aircraft.rssi = sig
			if !aircraft.mlat {
				aircraft.mlat = isMlat
			}
		}
		aircraft.lastPing = time.Now()
	}
	//fmt.Println(aircraft)
	//fmt.Println(aircraftExists)

	if df == 0 || df == 4 || df == 16 || df == 20 {
		m_bit := message[3] & (1<<6)
		q_bit := message[3] & (1<<4)

		if info.Debug {
			fmt.Printf("m_bit is %d, q_bit is %d\n", m_bit, q_bit)
		}

		aircraft.altUnit = 999
		if m_bit == 0 {
			aircraft.altUnit = 0 //Feet
			if (q_bit != 0) {
				/* N is the 11 bit integer resulting from the removal of bit Q and M */
				n := ((message[2]&31)<<6) |
					((message[3]&0x80)>>2) |
					((message[3]&0x20)>>1) |
					(message[3]&15);
				/* The final altitude is due to the resulting number multiplied by 25, minus 1000. */
				aircraft.altitude = (int32(n) * 25) - 1000;
			} else {
				/* TODO: Implement altitude where Q=0 and M=0 */
			}
		} else {
			aircraft.altUnit = 1
			/* TODO: Implement altitude when meter unit is selected. */
		}


/*
		altCode = (uint16(message[2])*256 + uint16(message[3])) & 0x1FFF

		if (altCode & 0x0040) > 0 {
			// meters
			// TODO
			//fmt.Println("meters")

		} else if (altCode & 0x0010) > 0 {
			// feet, raw integer
			ac := (altCode&0x1F80)>>2 + (altCode&0x0020)>>1 + (altCode & 0x000F)
			altitude = int32((ac * 25) - 1000)
			// TODO
			//fmt.Println("int altitude: ", altitude)

		} else if (altCode & 0x0010) == 0 {
			// feet, Gillham coded
			// TODO
			//fmt.Println("gillham")
		}

		if altitude != math.MaxInt32 {
			aircraft.altitude = altitude
		}
*/

	}

	if df == 17 || df == 18 {
		if info.Debug {
			spew.Dump(message)
		}
		decodeExtendedSquitter(message, df, &aircraft)
	}

	//logrus.Info("Returning Aircraft")
	if icaoAddr != math.MaxUint32 {
		//knownAircraft.Store(icaoAddr, &aircraft)
		return aircraft
	}
	return aircraftData{}
	//fmt.Println(aircraft)
}

func decodeModeAC(message []byte, isMlat bool, sig float64) aircraftData { //}, knownAircraft *AircraftMap) {
	// TODO
	return aircraftData{}
}

func parseTime(timebytes []byte) time.Time {
	// Takes a 6 byte array, which represents a 48bit GPS timestamp
	// http://wiki.modesbeast.com/Radarcape:Firmware_Versions#The_GPS_timestamp
	// and parses it into a Time.time

	upper := []byte{
		timebytes[0]<<2 + timebytes[1]>>6,
		timebytes[1]<<2 + timebytes[2]>>6,
		0, 0, 0, 0}
	lower := []byte{
		timebytes[2] & 0x3F, timebytes[3], timebytes[4], timebytes[5]}

	// the 48bit timestamp is 18bit day seconds | 30bit nanoseconds
	daySeconds := binary.BigEndian.Uint16(upper)
	nanoSeconds := int(binary.BigEndian.Uint32(lower))

	hr := int(daySeconds / 3600)
	min := int(daySeconds / 60 % 60)
	sec := int(daySeconds % 60)

	utcDate := time.Now().UTC()

	return time.Date(
		utcDate.Year(), utcDate.Month(), utcDate.Day(),
		hr, min, sec, nanoSeconds, time.UTC)
}

func decodeExtendedSquitter(message []byte, linkFmt uint, aircraft *aircraftData) {

	var callsign string

	if info.Debug {

		if linkFmt == 18 {
			switch message[0] & 7 {
			case 1:
				fmt.Println("Non-ICAO")
			case 2:
				fmt.Println("TIS-B fine")
			case 3:
				fmt.Println("TIS-B coarse")
			case 5:
				fmt.Println("TIS-B anon ADS-B relay")
			case 6:
				fmt.Println("ADS-B rebroadcast")
			default:
				fmt.Println("Non-ICAO unknown")
			}
		}
	}

	messageType := uint(message[4]) >> 3
	var msgSubType uint
	if messageType == 29 {
		msgSubType = (uint(message[4]) & 6) >> 1
	} else {
		msgSubType = uint(message[4]) & 7
	}

	//fmt.Printf("ext msg: %d\n", messageType)

	rawLatitude := uint32(math.MaxUint32)
	rawLongitude := uint32(math.MaxUint32)
	latitude := float64(math.MaxFloat64)
	longitude := float64(math.MaxFloat64)
	altitude := int32(math.MaxInt32)

	if len(message) < 10 {
		//spew.Dump(aircraft)
		logrus.Errorf(" Message length: %d\nMessage: %x\nMessage Type: %d\n Message Sub Type: %d", len(message), message, messageType, msgSubType)
	}

	if info.Debug {
		logrus.Debugf("Type: %d, Subtype %d Message: %x", messageType, msgSubType, message)
	}
	switch messageType {
	case 0:

	case 1, 2, 3, 4:

		callsign = parseCallsign(message)

		if info.Debug {
			logrus.Infof("Type %d (w/Callsign) %x", messageType, message)
		}

	case 19:
		if msgSubType >= 1 && msgSubType <= 4 {
			if msgSubType == 1 || msgSubType == 2 {
				aircraft.ewd = uint((message[5] & 4) >> 2)
				aircraft.ewv = int32(((message[5] & 3) << 8) | message[6])
				aircraft.nsd = uint((message[7] & 0x80) >> 7)
				aircraft.nsv = int32(((message[7] & 0x7f) << 3) | ((message[8] & 0xe0) >> 5))
				aircraft.vertRateSource = uint((message[8] & 0x10) >> 4)
				aircraft.vertRateSign = uint((message[8] & 0x8) >> 3)
				aircraft.vertRate = int32(((message[8] & 7) << 6) | ((message[9] & 0xfc) >> 2))

				aircraft.speed = int32(math.Sqrt(float64(aircraft.nsv*aircraft.nsv + aircraft.ewv*aircraft.ewv)))

				if aircraft.speed != 0 {
					ewv := aircraft.ewv
					nsv := aircraft.nsv

					if aircraft.ewd != 0 {
						ewv *= -1
					}
					if aircraft.nsd != 0 {
						nsv *= -1
					}

					heading := math.Atan2(float64(ewv), float64(nsv))

					aircraft.heading = int16((heading * 360) / (math.Pi * 2))
					if aircraft.heading < 0 {
						aircraft.heading += 360
					}
				} else {
					aircraft.heading = 0
				}
			} else if msgSubType == 3 || msgSubType == 4 {
				aircraft.headingIsValid = message[5]&(1<<2) != 0
				//aircraft.heading = uint16((360.0 / 128) * (((message[5] & 3) << 5) | (message[6] >> 3)))
			}
		}

	case 5, 6, 7, 8:
		// Ground position
		rawLatitude = uint32(message[6])&3<<15 + uint32(message[7])<<7 +
			uint32(message[8])>>1
		rawLongitude = uint32(message[8])&1<<16 + uint32(message[9])<<8 +
			uint32(message[10])

	case 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 20, 21, 22:
		// Airborne position
		/*		m_bit := message[5] & (1<<6)
				q_bit := message[6] & (1<<4)
		*/
		/*		if info.Debug {
					fmt.Printf("m_bit is %d, q_bit is %d\n", m_bit, q_bit)
				}
		*/
		ac12Data := (uint(message[5]) << 4) + (uint(message[6])>>4)&0x0FFF
		if messageType != 0 {
			rawLatitude = uint32(message[6])&3<<15 + uint32(message[7])<<7 +
				uint32(message[8])>>1
			rawLongitude = uint32(message[8])&1<<16 + uint32(message[9])<<8 +
				uint32(message[10])
		}
		if messageType != 20 && messageType != 21 && messageType != 22 {
			//altitude :=
			//fmt.Printf("ac12: %#04x\n", ac12Data)
			//fmt.Printf("ac12: %d\n", decodeAC12Field(ac12Data))

			altitude = decodeAC12Field(ac12Data)

		} else {
			// "HAE" ac2-encoded altitude
			// TODO
		}
	}

	if (rawLatitude != math.MaxUint32) && (rawLongitude != math.MaxUint32) {
		tFlag := (byte(message[6]) & 8) == 8
		isOddFrame := (byte(message[6]) & 4) == 4

		if isOddFrame && aircraft.eRawLat != math.MaxUint32 && aircraft.eRawLon != math.MaxUint32 {
			// Odd frame and we have previous even frame data
			latitude, longitude = parseRawLatLon(aircraft.eRawLat, aircraft.eRawLon, rawLatitude, rawLongitude, isOddFrame, tFlag)
			// Reset our buffer
			aircraft.eRawLat = math.MaxUint32
			aircraft.eRawLon = math.MaxUint32
		} else if !isOddFrame && aircraft.oRawLat != math.MaxUint32 && aircraft.oRawLon != math.MaxUint32 {
			// Even frame and we have previous odd frame data
			latitude, longitude = parseRawLatLon(rawLatitude, rawLongitude, aircraft.oRawLat, aircraft.oRawLon, isOddFrame, tFlag)
			// Reset buffer
			aircraft.oRawLat = math.MaxUint32
			aircraft.oRawLon = math.MaxUint32
		} else if isOddFrame {
			aircraft.oRawLat = rawLatitude
			aircraft.oRawLon = rawLongitude
		} else if !isOddFrame {
			aircraft.eRawLat = rawLatitude
			aircraft.eRawLon = rawLongitude
		}
	}

	switch msgSubType {
	case 1:
		break
	}

	if callsign != "" {
		aircraft.callsign = callsign
	}
	if altitude != math.MaxInt32 {
		aircraft.altitude = altitude
	}
	if latitude != math.MaxFloat64 && longitude != math.MaxFloat64 {
		aircraft.latitude = latitude
		aircraft.longitude = longitude
		aircraft.lastPos = time.Now()
	}
	if info.Debug {
		spew.Dump(aircraft)
	}
}

func parseCallsign(message []byte) string {
	var flight [8]byte
	flight[0] = aisChars[message[5]>>2]
	flight[1] = aisChars[((message[5]&3)<<4)|(message[6]>>4)]
	flight[2] = aisChars[((message[6]&15)<<2)|(message[7]>>6)]
	flight[3] = aisChars[message[7]&63]
	flight[4] = aisChars[message[8]>>2]
	flight[5] = aisChars[((message[8]&3)<<4)|(message[9]>>4)]
	flight[6] = aisChars[((message[9]&15)<<2)|(message[10]>>6)]
	flight[7] = aisChars[message[10]&63]
	return strings.TrimSpace(string(flight[:8]))
}

func parseRawLatLon(evenLat uint32, evenLon uint32, oddLat uint32,
	oddLon uint32, lastOdd bool, tFlag bool) (latitude float64, longitude float64) {
	if evenLat == math.MaxUint32 || oddLat == math.MaxUint32 || oddLon == math.MaxUint32 {
		return math.MaxFloat64, math.MaxFloat64
	}

	//fmt.Printf("Parsing: %d,%d + %d,%d\n", evenLat, evenLon, oddLat, oddLon)

	// http://www.lll.lu/~edward/edward/adsb/DecodingADSBposition.html
	j := int32((float64(59*evenLat-60*oddLat) / 131072.0) + 0.5)
	//fmt.Println("J: ", j)

	const airdlat0 = float64(6.0)
	const airdlat1 = float64(360.0) / float64(59.0)

	rlatEven := airdlat0 * (float64(j%60) + float64(evenLat)/131072.0)
	rlatOdd := airdlat1 * (float64(j%59) + float64(oddLat)/131072.0)
	if rlatEven >= 270 {
		rlatEven -= 360
	}
	if rlatOdd >= 270 {
		rlatOdd -= 360
	}

	//fmt.Println("rlat(0): ", rlatEven)
	//fmt.Println("rlat(1): ", rlatOdd)

	nlEven := cprnl(rlatEven)
	nlOdd := cprnl(rlatOdd)

	if nlEven != nlOdd {
		return math.MaxFloat64, math.MaxFloat64
	}

	//fmt.Println("NL(0): ", nlEven)
	//fmt.Println("NL(1): ", nlOdd)

	var ni int16

	if lastOdd {
		ni = int16(nlOdd) - 1
	} else {
		ni = int16(nlEven) - 1
	}
	if ni < 1 {
		ni = 1
	}
	//fmt.Println("NL(i): ", ni)

	//dlon := 360.0/float64(ni)
	//fmt.Println("dlon(i):", dlon)

	var m int16
	var outLat float64
	var outLon float64
	if tFlag {
		m = int16(math.Floor((float64(int32(evenLon*uint32(cprnl(rlatOdd)-1))-
			int32(oddLon*uint32(cprnl(rlatOdd)))) / 131072.0) + 0.5))
		outLon = cprDlonFunction(rlatOdd, tFlag, false) * (float64(m%ni) + float64(oddLon)/131072.0)
		outLat = rlatOdd

	} else {
		m = int16(math.Floor((float64(int32(evenLon*uint32(cprnl(rlatEven)-1))-
			int32(oddLon*uint32(cprnl(rlatEven)))) / 131072.0) + 0.5))
		outLon = cprDlonFunction(rlatEven, tFlag, false) * (float64(m%ni) + float64(evenLon)/131072.0)
		outLat = rlatEven
	}

	outLon -= math.Floor((outLon+180.0)/360.0) * 360.0

	//fmt.Println("M: ", m)
	//fmt.Println("outLat: ", outLat)
	//fmt.Println("outLon: ", outLon)

	return outLat, outLon
}

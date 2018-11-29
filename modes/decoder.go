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
	"github.com/ccustine/beastie/config"
	"github.com/ccustine/beastie/types"
	"github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"strings"

	//"fmt"
	"math"
	"time"
)

const (
	aisChars             = "@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_ !\"#$%&'()*+,-./0123456789:;<=>?"
	MODES_GENERATOR_POLY = uint32(0xfff409)
)

var (
	crcTable [256]uint32
)

func init() {
	var i uint32

	for i = 0; i < 256; i += 1 {

		var c = i << 16

		var j uint
		for j = 0; j < 8; j += 1 {
			if (c & 0x800000) != 0 {
				c = (c << 1) ^ MODES_GENERATOR_POLY
			} else {
				c = (c << 1)
			}
		}
		crcTable[i] = c & 0x00ffffff
	}
}

func DecodeModeS(message []byte, isMlat bool, sig float64, knownAircraft *types.AircraftMap, info *config.BeastInfo) types.AircraftData {
	df := getbits(message, 1, 5) //uint((message[0] & 0xF8) >> 3)

	var contextLogger *log.Entry
	if info.Debug {
		contextLogger = log.WithFields(log.Fields{
			"message": fmt.Sprintf("%x", message),
			"Mlat":    isMlat,
			"Rssi":    sig,
		})
	}

	var aircraft types.AircraftData
	//var aircraftExists bool
	//aircraft.VertRateSign = math.MaxUint32
	icaoAddr := uint32(math.MaxUint32)
	squawk := uint32(0)
	//altCode := uint16(math.MaxUint16)
	//Altitude := int32(math.MaxInt32)

	var msgType string

	metrics.GetOrRegisterCounter(fmt.Sprintf("DF %02d", df), nil).Inc(1)

	switch df {
	case 0:
		msgType = "short air-air surveillance (TCAS)"
	case 4:
		msgType = "surveillance, Altitude reply"
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
		msgType = "Comm-B including Altitude reply"
	case 21:
		msgType = "Comm-B reply including Mode A identity"
	case 22:
		msgType = "military use"
	case 24:
		msgType = "special long msg"
	default:
		msgType = "unknown"
	}

	//if df == 0 || df == 4 || df == 11 || df == 17 || df == 18 {
	if df == 11 || df == 17 || df == 18 {
		icaoAddr = getbits(message, 9, 32) //uint32(message[1])<<16 | uint32(message[2])<<8 | uint32(message[3]) //New
		/*		if info.Debug {
					contextLogger.WithField("icao", fmt.Sprintf("ICAO: %06x\n", IcaoAddr))
					//log.Debugf("ICAO: %06x\n", IcaoAddr)
				}
		*/
	}

	if df == 5 || df == 21 {
		var bits uint
		if df == 5 {
			bits = 56
		} else if df == 21 {
			bits = 112
		}
		icaoAddr = modesChecksum(message, bits)

		id := getbits(message, 20, 32)
		squawk = uint32(decodeID13Field(uint(id)))
		if info.Debug {
			log.Debugf("Ident Msg: %x Squawk code: %04x id %d, ICAO: %x", message, squawk, id, icaoAddr)
		}
	}

	if icaoAddr != math.MaxUint32 {
		ptrAircraft, aircraftExists := knownAircraft.Load(icaoAddr)
		if !aircraftExists {
			// Initial values
			aircraft = types.AircraftData{
				IcaoAddr:     icaoAddr,
				Squawk:       squawk,
				ORawLat:      math.MaxUint32,
				ORawLon:      math.MaxUint32,
				ERawLat:      math.MaxUint32,
				ERawLon:      math.MaxUint32,
				Latitude:     math.MaxFloat64,
				Longitude:    math.MaxFloat64,
				Altitude:     math.MaxInt32,
				Callsign:     "",
				Mlat:         isMlat,
				Rssi:         sig,
				VertRateSign: math.MaxUint32,
				IsValid:      true,
			}
		} else {
			aircraft = *ptrAircraft
			aircraft.Rssi = sig
			if !aircraft.Mlat {
				aircraft.Mlat = isMlat
			}
			if squawk != 0 {
				aircraft.Squawk = squawk
			}
		}
		aircraft.LastPing = time.Now()
	} else {
		return types.AircraftData{IsValid: false}

	}
	//log.Debugf(aircraft)
	//log.Debugf(aircraftExists)

	if df == 0 || df == 4 || df == 16 || df == 20 {
		m_bit := message[3] & (1 << 6)
		q_bit := message[3] & (1 << 4)

		/*		if info.Debug {
					log.Debugf("m_bit is %d, q_bit is %d\n", m_bit, q_bit)
				}
		*/
		//var altUnit = 999
		if m_bit == 0 {
			//altUnit = 0 //Feet
			if (q_bit != 0) {
				/* N is the 11 bit integer resulting from the removal of bit Q and M */
				n := ((message[2] & 31) << 6) |
					((message[3] & 0x80) >> 2) |
					((message[3] & 0x20) >> 1) |
					(message[3] & 15);
				/* The final Altitude is due to the resulting number multiplied by 25, minus 1000. */
				aircraft.Altitude = (int32(n) * 25) - 1000;
			} else {
				/* TODO: Implement Altitude where Q=0 and M=0 */
			}
		} else {
			//altUnit = 1
			/* TODO: Implement Altitude when meter unit is selected. */
		}

		/*
				altCode = (uint16(message[2])*256 + uint16(message[3])) & 0x1FFF

				if (altCode & 0x0040) > 0 {
					// meters
					// TODO
					//log.Debugf("meters")

				} else if (altCode & 0x0010) > 0 {
					// feet, raw integer
					ac := (altCode&0x1F80)>>2 + (altCode&0x0020)>>1 + (altCode & 0x000F)
					Altitude = int32((ac * 25) - 1000)
					// TODO
					//log.Debugf("int Altitude: ", Altitude)

				} else if (altCode & 0x0010) == 0 {
					// feet, Gillham coded
					// TODO
					//log.Debugf("gillham")
				}

				if Altitude != math.MaxInt32 {
					aircraft.Altitude = Altitude
				}
		*/

	}

	if df == 17 || df == 18 {
		//if info.Debug {
		//	spew.Dump(message)
		//}
		if len(message) != 14 {
			log.Debug("ES Message was not 14 bytes: %x", message)
			// TODO: Maybe need to return empty aircraft here?
		} else {
			DecodeExtendedSquitter(message, uint(df), &aircraft, info)
		}
	}

	//log.Info("Returning Aircraft")
	if icaoAddr != math.MaxUint32 {
		if info.Debug {
			contextLogger = log.WithFields(log.Fields{
				"message": fmt.Sprintf("%x", message),
				"Mlat":    isMlat,
				"Rssi":    sig,
				"icao":    icaoAddr,
				"msgtype": msgType,
			})
		}

		return aircraft
	}
	if info.Debug {
		contextLogger.Debugf("Returning empty aircraft")
	}
	return types.AircraftData{IsValid: false}
	//log.Debugf(aircraft)
}

func DecodeModeAC(message []byte, isMlat bool, sig float64, knownAircraft *types.AircraftMap, info *config.BeastInfo) types.AircraftData {
	// TODO
	if info.Debug {
		log.Debug("NOOP on ModeAC decode")
	}
	return types.AircraftData{}
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

func DecodeExtendedSquitter(message []byte, linkFmt uint, aircraft *types.AircraftData, info *config.BeastInfo) {

	var callsign string

	if info.Debug {

		if linkFmt == 18 {
			switch message[0] & 7 {
			case 1:
				log.Debugf("ES Non-ICAO")
			case 2:
				log.Debugf("ES TIS-B fine")
			case 3:
				log.Debugf("ES TIS-B coarse")
			case 5:
				log.Debugf("ES TIS-B anon ADS-B relay")
			case 6:
				log.Debugf("ES ADS-B rebroadcast")
			default:
				log.Debugf("ES Non-ICAO unknown")
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

	//log.Debugf("ext msg: %d\n", messageType)

	rawLatitude := uint32(math.MaxUint32)
	rawLongitude := uint32(math.MaxUint32)
	latitude := float64(math.MaxFloat64)
	longitude := float64(math.MaxFloat64)
	altitude := int32(math.MaxInt32)

	if len(message) < 10 {
		//spew.Dump(aircraft)
		log.Errorf(" Message length: %d\nMessage: %x\nMessage Type: %d\n Message Sub Type: %d", len(message), message, messageType, msgSubType)
	}

	/*	if info.Debug {
			log.Debugf("Type: %d, Subtype %d Message: %x", messageType, msgSubType, message)
		}
	*/
	switch messageType {
	case 0:

	case 1, 2, 3, 4:

		callsign = parseCallsign(message)

		/*		if info.Debug {
					log.Infof("Type %d (w/Callsign) %x", messageType, message)
				}
		*/
	case 19:
		if msgSubType >= 1 && msgSubType <= 4 {
			if msgSubType == 1 || msgSubType == 2 {
				ewd := int32((message[5] & 4) >> 2)
				ewv := (int32(message[5])&3)<<8 | int32(message[6])
				nsd := int32((message[7] & 0x80) >> 7)
				nsv := (int32(message[7])&0x7f)<<3 | (int32(message[8])&0xe0)>>5
				aircraft.VertRateSource = uint((message[8] & 0x10) >> 4)
				aircraft.VertRateSign = uint((message[8] & 0x8) >> 3)
				aircraft.VertRate = int32(math.Round(float64(((int32((message[8]&7)<<6|(message[9]&0xfc)>>2)-1)*64)/25) * 25))

				aircraft.Speed = int32(math.Sqrt(float64(nsv*nsv + ewv*ewv)))

				if aircraft.Speed != 0 {
					if ewd != 0 {
						ewv *= -1
					}
					if nsd != 0 {
						nsv *= -1
					}

					heading := math.Atan2(float64(ewv), float64(nsv))

					aircraft.Heading = int32((heading * 360) / (math.Pi * 2))
					if aircraft.Heading < 0 {
						aircraft.Heading += 360
					}
				} else {
					aircraft.Heading = 0
				}
			} else if msgSubType == 3 || msgSubType == 4 {
				aircraft.HeadingIsValid = message[5]&(1<<2) != 0
				aircraft.Heading = int32(math.Round(360.0/128)) * (((int32(message[5]) & 3) << 5) | (int32(message[6]) >> 3))
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
					log.Debugf("m_bit is %d, q_bit is %d\n", m_bit, q_bit)
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
			//Altitude :=
			//log.Debugf("ac12: %#04x\n", ac12Data)
			//log.Debugf("ac12: %d\n", decodeAC12Field(ac12Data))

			altitude = decodeAC12Field(ac12Data)

		} else {
			// "HAE" ac2-encoded Altitude
			// TODO
		}
	}

	if (rawLatitude != math.MaxUint32) && (rawLongitude != math.MaxUint32) {
		tFlag := (byte(message[6]) & 8) == 8
		isOddFrame := (byte(message[6]) & 4) == 4

		if isOddFrame && aircraft.ERawLat != math.MaxUint32 && aircraft.ERawLon != math.MaxUint32 {
			// Odd frame and we have previous even frame data
			latitude, longitude = parsERawLatLon(aircraft.ERawLat, aircraft.ERawLon, rawLatitude, rawLongitude, isOddFrame, tFlag)
			// Reset our buffer
			aircraft.ERawLat = math.MaxUint32
			aircraft.ERawLon = math.MaxUint32
		} else if !isOddFrame && aircraft.ORawLat != math.MaxUint32 && aircraft.ORawLon != math.MaxUint32 {
			// Even frame and we have previous odd frame data
			latitude, longitude = parsERawLatLon(rawLatitude, rawLongitude, aircraft.ORawLat, aircraft.ORawLon, isOddFrame, tFlag)
			// Reset buffer
			aircraft.ORawLat = math.MaxUint32
			aircraft.ORawLon = math.MaxUint32
		} else if isOddFrame {
			aircraft.ORawLat = rawLatitude
			aircraft.ORawLon = rawLongitude
		} else if !isOddFrame {
			aircraft.ERawLat = rawLatitude
			aircraft.ERawLon = rawLongitude
		}
	}

	switch msgSubType {
	case 1:
		break
	}

	if callsign != "" {
		aircraft.Callsign = callsign
	}
	if altitude != math.MaxInt32 {
		aircraft.Altitude = altitude
	}
	if latitude != math.MaxFloat64 && longitude != math.MaxFloat64 {
		aircraft.Latitude = latitude
		aircraft.Longitude = longitude
		aircraft.LastPos = time.Now()
	}
	//if info.Debug {
	//	spew.Dump(aircraft)
	//}
}

func decodeID13Field(ID13Field uint) uint {
	var hexGillham uint = 0

	if ID13Field&0x1000 != 0 {
		hexGillham |= 0x0010
	} // Bit 12 = C1
	if ID13Field&0x0800 != 0 {
		hexGillham |= 0x1000
	} // Bit 11 = A1
	if ID13Field&0x0400 != 0 {
		hexGillham |= 0x0020
	} // Bit 10 = C2
	if ID13Field&0x0200 != 0 {
		hexGillham |= 0x2000
	} // Bit  9 = A2
	if ID13Field&0x0100 != 0 {
		hexGillham |= 0x0040
	} // Bit  8 = C4
	if ID13Field&0x0080 != 0 {
		hexGillham |= 0x4000
	} // Bit  7 = A4
	//if (ID13Field & 0x0040) {hexGillham |= 0x0800;} // Bit  6 = X  or M
	if ID13Field&0x0020 != 0 {
		hexGillham |= 0x0100
	} // Bit  5 = B1
	if ID13Field&0x0010 != 0 {
		hexGillham |= 0x0001
	} // Bit  4 = D1 or Q
	if ID13Field&0x0008 != 0 {
		hexGillham |= 0x0200
	} // Bit  3 = B2
	if ID13Field&0x0004 != 0 {
		hexGillham |= 0x0002
	} // Bit  2 = D2
	if ID13Field&0x0002 != 0 {
		hexGillham |= 0x0400
	} // Bit  1 = B4
	if ID13Field&0x0001 != 0 {
		hexGillham |= 0x0004
	} // Bit  0 = D4

	return hexGillham
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

func parsERawLatLon(evenLat uint32, evenLon uint32, oddLat uint32,
	oddLon uint32, lastOdd bool, tFlag bool) (latitude float64, longitude float64) {
	if evenLat == math.MaxUint32 || oddLat == math.MaxUint32 || oddLon == math.MaxUint32 {
		return math.MaxFloat64, math.MaxFloat64
	}

	//log.Debugf("Parsing: %d,%d + %d,%d\n", evenLat, evenLon, oddLat, oddLon)

	// http://www.lll.lu/~edward/edward/adsb/DecodingADSBposition.html
	j := int32((float64(59*evenLat-60*oddLat) / 131072.0) + 0.5)
	//log.Debugf("J: ", j)

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

	//log.Debugf("rlat(0): ", rlatEven)
	//log.Debugf("rlat(1): ", rlatOdd)

	nlEven := cprnl(rlatEven)
	nlOdd := cprnl(rlatOdd)

	if nlEven != nlOdd {
		return math.MaxFloat64, math.MaxFloat64
	}

	//log.Debugf("NL(0): ", nlEven)
	//log.Debugf("NL(1): ", nlOdd)

	var ni int16

	if lastOdd {
		ni = int16(nlOdd) - 1
	} else {
		ni = int16(nlEven) - 1
	}
	if ni < 1 {
		ni = 1
	}
	//log.Debugf("NL(i): ", ni)

	//dlon := 360.0/float64(ni)
	//log.Debugf("dlon(i):", dlon)

	var m int16
	var outLat float64
	var outLon float64
	if tFlag {
		m = int16(math.Floor((float64(int32(evenLon*uint32(cprnl(rlatOdd)-1)) -
			int32(oddLon*uint32(cprnl(rlatOdd)))) / 131072.0) + 0.5))
		outLon = cprDlonFunction(rlatOdd, tFlag, false) * (float64(m%ni) + float64(oddLon)/131072.0)
		outLat = rlatOdd

	} else {
		m = int16(math.Floor((float64(int32(evenLon*uint32(cprnl(rlatEven)-1)) -
			int32(oddLon*uint32(cprnl(rlatEven)))) / 131072.0) + 0.5))
		outLon = cprDlonFunction(rlatEven, tFlag, false) * (float64(m%ni) + float64(evenLon)/131072.0)
		outLat = rlatEven
	}

	outLon -= math.Floor((outLon+180.0)/360.0) * 360.0

	//log.Debugf("M: ", m)
	//log.Debugf("outLat: ", outLat)
	//log.Debugf("outLon: ", outLon)

	return outLat, outLon
}

func getbits(data []byte, firstbit uint16, lastbit uint16) uint32 {
	fbi := firstbit - 1
	lbi := lastbit - 1
	//nbi := lastbit - firstbit + 1

	fby := uint16(fbi >> 3)
	lby := uint16(lbi >> 3)
	nby := uint16((lby - fby) + 1)

	shift := 7 - (lbi & 7)
	topmask := uint(0xFF >> (fbi & 7))

	//assert (fbi <= lbi);
	//assert (nbi <= 32);
	//assert (nby <= 5);

	//log.Infof("nby is %d", nby)

	if (nby == 5) {
		return uint32((uint(data[fby])&topmask)<<(32-shift) |
			uint(data[fby+1])<<(24-shift) |
			uint(data[fby+2])<<(16-shift) |
			uint(data[fby+3])<<(8-shift) |
			uint(data[fby+4])>>shift)
	} else if (nby == 4) {
		return uint32((uint(data[fby])&topmask)<<(24-shift) |
			uint(data[fby+1])<<(16-shift) |
			uint(data[fby+2])<<(8-shift) |
			uint(data[fby+3])>>shift)
	} else if (nby == 3) {
		return uint32((uint(data[fby])&topmask)<<(16-shift) |
			uint(data[fby+1])<<(8-shift) |
			uint(data[fby+2])>>shift)
	} else if (nby == 2) {
		return uint32((uint(data[fby])&topmask)<<(8-shift) |
			uint(data[fby+1])>>shift)
	} else if (nby == 1) {
		return uint32(uint(data[fby])&topmask) >> shift
	} else {
		return 0
	}
}

func modesChecksum(message []byte, bits uint) uint32 {
	var rem uint32 = 0
	var i uint

	n := bits / 8

	//assert(bits%8 == 0)
	//assert(n >= 3)

	for i = 0; i < n-3; i = i + 1 {
		//log.Debugf("Checksum message %x", message)
		//log.Debugf("Checksum index %d, n is %d", i, n)
		//log.Debugf("Checksum calc %d", uint32(message[i])^((rem&0xff0000)>>16))
		rem = (rem << 8) ^ crcTable[uint32(message[i])^((rem&0xff0000)>>16)]
		rem = rem & 0xffffff
	}

	rem = rem ^ (uint32(message[n-3]) << 16) ^ (uint32(message[n-2]) << 8) ^ (uint32(message[n-1]))
	return rem
}

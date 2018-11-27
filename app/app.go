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

package app

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/SierraSoftworks/multicast"
	. "github.com/ccustine/beastie/config"
	"github.com/ccustine/beastie/modes"
	"github.com/ccustine/beastie/output"
	"github.com/ccustine/beastie/types"
	"github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"math"
	"net"
	"time"
)

var (
	magicTimestampMLAT = []byte{0xFF, 0x00, 0x4D, 0x4C, 0x41, 0x54}
	Info               BeastInfo
	knownAircraft      = types.NewAircraftMap()
	aircraft           = make(chan types.AircraftData) //, 10)
	GoodRate           = metrics.GetOrRegisterMeter("Message Rate (Good)", metrics.DefaultRegistry)
	BadRate            = metrics.GetOrRegisterMeter("Message Rate (Bad)", metrics.DefaultRegistry)
	ModeACCnt          = metrics.GetOrRegisterCounter("Message Rate (ModeA/C)", metrics.DefaultRegistry)
	ModesShortCnt      = metrics.GetOrRegisterCounter("Message Rate (ModeS Short)", metrics.DefaultRegistry)
	ModesLongCnt       = metrics.GetOrRegisterCounter("Message Rate (ModeS Long)", metrics.DefaultRegistry)
)

type TCPClient struct {
	Host string
	Port int
}

func (c *TCPClient) start(ac chan types.AircraftData) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		panic(err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Error(err)
	}

	go handleConnection(conn, ac)
}

func Start(beastInfo BeastInfo) {
	Info = beastInfo
	mcast := multicast.New()

	sources := make(map[string]*TCPClient)
	if beastInfo.Debug {
		log.Debugf("Beast Info: %v", beastInfo)
	}

	for _, source := range beastInfo.Sources {
		sourceKey := fmt.Sprintf("%s:%d", source.Host, source.Port)
		if source.Host != "" && source.Port != 0 {
			sources[sourceKey] = &TCPClient{
				Host: source.Host,
				Port: source.Port,
			}
			sources[sourceKey].start(aircraft)
		}
	}

	// TODO: Add a map or array to store generic Outputs interface types
	// and access using the interface methods

	var outputs = make([]interface{}, len(beastInfo.Outputs))
	for i, outtype := range beastInfo.Outputs {
		switch outtype {
		case output.TABLE:
			outputs[i] = output.NewTableOutput(&beastInfo)
		case output.LOG:
			outputs[i] = output.NewLogOutput(&beastInfo)
		case output.JSONAPI:
			outputs[i] = output.NewJsonOutput()
		default:
			outputs[i] = output.NewTableOutput(&beastInfo)
		}
		go func(op output.Output) {
			l := mcast.Listen()
			for {
				acm := <-l.C
				op.UpdateDisplay(acm.(types.AircraftMap))
			}
		}(outputs[i].(output.Output))
	}

	go func() {
		ticker := time.NewTicker(1000 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				evict := false
				for _, aircraft := range knownAircraft.Range() {
					if !aircraft.LastPing.IsZero() {
						evict = time.Since(aircraft.LastPing) > (time.Duration(59) * time.Second)
					}

					if evict {
						knownAircraft.Delete(aircraft.IcaoAddr)
						continue
					}
				}
				mcast.C <- knownAircraft
			}
		}
	}()

	for {
		select {
		case airframe := <-aircraft:
			if !airframe.IsValid {
				continue
			}
			knownAircraft.Store(airframe.IcaoAddr, &airframe)
		}
	}

}

func handleConnection(conn net.Conn, ac chan types.AircraftData) {
	//reader := bufio.NewReaderSize(conn, 128)
	reader := bufio.NewReader(conn)
	scanner := bufio.NewScanner(reader)
	scanner.Split(ScanModeS)

	defer conn.Close()

	for scanner.Scan() {
		currentMessage := scanner.Bytes()

		// Connection closed
		if len(currentMessage) == 0 {
			continue
		}

		validMessage := false
		if currentMessage[0] == 0x31 || currentMessage[0] == 0x32 ||
			currentMessage[0] == 0x33 || currentMessage[0] == 0x34 {
			validMessage = true
		}
		if !validMessage {
			if Info.Debug {
				log.Debugf("Not a valid Message with 0x31 32 33 34 Msg: %#x\n", currentMessage)
			}
			continue
		}

		msgType := currentMessage[0]
		var msgLen int

		// http://wiki.modesbeast.com/Mode-S_Beast:Data_Output_Formats
		switch msgType {
		case 0x31: // 1 - Mode A/C
			ModeACCnt.Inc(1)
			msgLen = 10
			if Info.Debug {
				log.Debugf("Invalid Beast mode msg type 1: %x", currentMessage)
			}
		case 0x32: // 2 - Mode S Short
			ModesShortCnt.Inc(1)
			msgLen = 15
		case 0x33: // 3 - Mode S Long
			ModesLongCnt.Inc(1)
			msgLen = 22
		case 0x34: // 4
			if (Info.Debug) {
				log.Debugf("Invalid Beast mode msg type 4: %x", currentMessage)
			}
			continue // not supported
		default:
			continue
			//msgLen = 8 // shortest possible msg w/header & timetstamp
		}

		if len(currentMessage) == msgLen {
			GoodRate.Mark(1)
		} else {
			BadRate.Mark(1)
			continue
		}

		isMlat := bytes.Equal(currentMessage[1:7], magicTimestampMLAT)

		if msgType == 0x31 {
			ac <- modes.DecodeModeAC(currentMessage[8:], isMlat, 10*math.Log10(math.Pow(float64(currentMessage[7])/255, 2)), &knownAircraft, &Info)
		} else {
			ac <- modes.DecodeModeS(currentMessage[8:], isMlat, 10*math.Log10(math.Pow(float64(currentMessage[7])/255, 2)), &knownAircraft, &Info)
		}
	}

	if scanner.Err() != nil {
		if Info.Debug {
			log.Debugf("Scanner Error: %s\n", scanner.Err())
		}
	}

}

func ScanModeS(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0x1a); i >= 0 {
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

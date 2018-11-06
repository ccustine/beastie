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
	"bufio"
	"bytes"
	"fmt"
	"github.com/SierraSoftworks/multicast"
	. "github.com/ccustine/beastie/beastie"
	"github.com/ccustine/uilive"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"math"
	"net"
	"reflect"
	"time"
)

var (
	magicTimestampMLAT = []byte{0xFF, 0x00, 0x4D, 0x4C, 0x41, 0x54}
	uiWriter           *uilive.Writer
	info               BeastInfo
	knownAircraft      = NewAircraftMap()
	aircraft           = make(chan aircraftData)
	GoodRate           = metrics.GetOrRegisterMeter("Message Rate (Good)", metrics.DefaultRegistry)
	BadRate            = metrics.GetOrRegisterMeter("Message Rate (Bad)", metrics.DefaultRegistry)
	ModeACCnt          = metrics.GetOrRegisterCounter("Message Rate (ModeA/C)", metrics.DefaultRegistry)
	ModesShortCnt      = metrics.GetOrRegisterCounter("Message Rate (ModeS Short)", metrics.DefaultRegistry)
	ModesLongCnt       = metrics.GetOrRegisterCounter("Message Rate (ModeS Long)", metrics.DefaultRegistry)
)

const (
	aisChars = "@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_ !\"#$%&'()*+,-./0123456789:;<=>?"
)

type TCPClient struct {
	Host string
	Port int
}

func (c *TCPClient) start(ac chan aircraftData) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		panic(err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	handleConnection(conn, ac)
}

func Stop() {

}

func Start(beastInfo BeastInfo) {
	info = beastInfo
	output := multicast.New()

	sources := make(map[string]*TCPClient)
	if beastInfo.Debug {
		logrus.Debugf("Beast Info: %v", beastInfo)
	}

	for _, source := range beastInfo.Sources {
		// TODO: This map and key is super fucking ugly, pls fixme
		sourceKey := fmt.Sprintf("%s:%d", source.Host, source.Port)
		if source.Host != "" && source.Port != 0 {
			sources[sourceKey] = &TCPClient{
				Host: source.Host,
				Port: source.Port,
			}
			go sources[sourceKey].start(aircraft)
		}
	}

	uiWriter = uilive.New()
	uiWriter.RefreshInterval = 1000 * time.Millisecond
	uiWriter.Start()

	go func() {
		l := output.Listen()
		for acm := range l.C {
			newac := acm.(*AircraftMap)
			updateDisplay(newac, uiWriter)
		}
	}()

	go func() {
		ticker := time.NewTicker(1000 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				output.C <- knownAircraft

				//updateDisplay(knownAircraft, uiWriter)

				//_ = uiWriter.Flush()
			}
		}
	}()

	for {
		select {
		case airframe := <-aircraft:
			knownAircraft.Store(airframe.icaoAddr, &airframe)
		}
	}

}

func handleConnection(conn net.Conn, ac chan aircraftData) {
	reader := bufio.NewReader(conn)
	scanner := bufio.NewScanner(reader)
	scanner.Split(ScanModeS)

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
			if info.Debug {
				logrus.Debugf("Not a valid Message with 0x31 32 33 34 Msg: %#x\n", currentMessage)
			}
			continue
		}

		message := currentMessage

		msgType := message[0]
		var msgLen int

		// http://wiki.modesbeast.com/Mode-S_Beast:Data_Output_Formats
		switch msgType {
		case 0x31: // 1
			ModeACCnt.Inc(1)
			msgLen = 10
			if (info.Debug) {
				logrus.Debugf("Invalid Beast mode msg type 1: %x", message)
			}
		case 0x32: // 2
			ModesShortCnt.Inc(1)
			msgLen = 15
		case 0x33: // 3
			ModesLongCnt.Inc(1)
			msgLen = 22
		case 0x34: // 4
			if (info.Debug) {
				logrus.Debugf("Invalid Beast mode msg type 4: %x", message)
			}
			continue // not supported
		default:
			continue
			//msgLen = 8 // shortest possible msg w/header & timetstamp
		}

		if len(message) == msgLen {
			logrus.Debugf("Message (Exact) len %d expected %d : %x", len(message), msgLen, message)
			// Mark the rate because now we are going to actually parse the message
			GoodRate.Mark(1)
		} else if len(message) <= msgLen {
			logrus.Debugf("Message (Less) len %d expected %d : %x", len(message), msgLen, message)
			BadRate.Mark(1)
			continue
		} else if len(message) > msgLen {
			logrus.Debugf("Message (More) len %d expected %d : %x", len(message), msgLen, message)
			BadRate.Mark(1)
			continue
		}

		isMlat := reflect.DeepEqual(message[1:7], magicTimestampMLAT)
		if isMlat {
			//fmt.Println("FROM MLAT")
			//timestamp := parseTime(message[1:7])
			//fmt.Println(otimestamp)
			//timestamp = time.Now()
		} else {
			//timestamp = parseTime(message[1:7])
			//_ = timestamp
			//fmt.Println(timestamp)
		}

		sigLevel := message[7]

		msgContent := message[8:]
		if info.Debug {
			fmt.Printf("%d byte frame\n", len(msgContent))
		}
		for i := 0; i < len(msgContent); i++ {
			if info.Debug {
				fmt.Printf("%02x", msgContent[i])
			}
		}
		if info.Debug {
			fmt.Printf("\n")
		}

		if msgType == 0x31 {
			ac <- decodeModeAC(msgContent, isMlat, 10*math.Log10(math.Pow(float64(sigLevel)/255, 2)))
		} else {
			ac <- decodeModeS(msgContent, isMlat, 10*math.Log10(math.Pow(float64(sigLevel)/255, 2)))
		}
	}

	if scanner.Err() != nil {
		if info.Debug {
			logrus.Debugf("Scanner Error: %s\n", scanner.Err())
		}
	}

}

func ScanModeS(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		//fmt.Sprintln("At EOF")
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0x1a); i >= 0 {
		//fmt.Sprintf("Found delim at %d, data is %#x\n", i, data)
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		//fmt.Sprintf("At EOF with final token %#x\n", data)
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

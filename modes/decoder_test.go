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
	"encoding/hex"
	"reflect"
	"testing"
	"time"
)

func Test_decodeModeS(t *testing.T) {
	type args struct {
		message []byte
		isMlat  bool
	}

	tests := []struct {
		name string
		args args
		want aircraftData
	}{
		{
			name: "Good test",
			args: args{convertToBytes("8dad73a999117b9b8004285d1c83"),
				false},
			want: aircraftData{callsign: "AAL2748 "},
		},
		{
			name: "Callsign test",
			args: args{convertToBytes("8da6c6c820053074db08208391f5"),
				false},
			want: aircraftData{icaoAddr:0xa6c6c8, callsign:"ASA460", eRawLat:0xffffffff, eRawLon:0xffffffff,
				oRawLat:0xffffffff, oRawLon:0xffffffff, latitude:1.7976931348623157e+308,
				longitude:1.7976931348623157e+308, altitude:2147483647, altUnit:0x0, ewd:0x0, ewv:0,
				nsd: 0x0, nsv:0, vertRateSource:0x0, vertRateSign:0x0, vertRate:0, speed:0, heading:0,
				headingIsValid:false, mlat:false},
		},
		{
			name: "Callsign test 2 index out of range",
			args: args{convertToBytes("95e51fbf0ef3e3ba"),
				false},
			want: aircraftData{icaoAddr:0xa6c6c8, callsign:"ASA460", eRawLat:0xffffffff, eRawLon:0xffffffff,
				oRawLat:0xffffffff, oRawLon:0xffffffff, latitude:1.7976931348623157e+308,
				longitude:1.7976931348623157e+308, altitude:2147483647, altUnit:0x0, ewd:0x0, ewv:0,
				nsd: 0x0, nsv:0, vertRateSource:0x0, vertRateSign:0x0, vertRate:0, speed:0, heading:0,
				headingIsValid:false, lastPing:createTime("2018-10-24 08:20:10.827814 -0600 MDT m=+42.649475709"),
				lastPos:createTime("2018-10-24 08:20:05.283828 -0600 MDT m=+37.105283228"), mlat:false},
		},
		{
			name: "Too Short",
			args: args{[]byte("8d4285d1c83"),
				false},
			want: aircraftData{},
		},
		{
			name: "Bad CA",
			args: args{[]byte("aaa8d4285d1c83"),
				false},
			want: aircraftData{},
		},
		{
			name: "Bad test 3",
			args: args{[]byte("8dad73a999117b9b8004285d"),
				false},
			want: aircraftData{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeModeS(tt.args.message, tt.args.isMlat, 0); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeModeS() = \ngot:  %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}

func Test_parseTime(t *testing.T) {
	type args struct {
		timebytes []byte
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTime(tt.args.timebytes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decodeExtendedSquitter(t *testing.T) {
	type args struct {
		message  []byte
		linkFmt  uint
		aircraft *aircraftData
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decodeExtendedSquitter(tt.args.message, tt.args.linkFmt, tt.args.aircraft)
		})
	}
}

func Test_parsRawLatLon(t *testing.T) {
	type args struct {
		evenLat uint32
		evenLon uint32
		oddLat  uint32
		oddLon  uint32
		lastOdd bool
		tFlag   bool
	}
	tests := []struct {
		name          string
		args          args
		wantLatitude  float64
		wantLongitude float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLatitude, gotLongitude := parseRawLatLon(tt.args.evenLat, tt.args.evenLon, tt.args.oddLat, tt.args.oddLon, tt.args.lastOdd, tt.args.tFlag)
			if gotLatitude != tt.wantLatitude {
				t.Errorf("parseRawLatLon() gotLatitude = %v, want %v", gotLatitude, tt.wantLatitude)
			}
			if gotLongitude != tt.wantLongitude {
				t.Errorf("parseRawLatLon() gotLongitude = %v, want %v", gotLongitude, tt.wantLongitude)
			}
		})
	}
}

func Test_parseCallsign(t *testing.T) {
	type args struct {
		message []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Good Callsign Test 1", args{convertToBytes("8dabeb31204d7074db782012f83a")}, "SWA467"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseCallsign(tt.args.message); got != tt.want {
				t.Errorf("parseCallsign() = %v, want %v", got, tt.want)
			}
		})
	}
}

func convertToBytes(from string) []byte {
	to, _ := hex.DecodeString(from)
	return to
}

func createTime(from string) time.Time {
	to, _ := time.Parse(time.RFC3339, from)
	return to
}
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
	"encoding/json"
	"fmt"
	_ "github.com/ccustine/beastie/statik"
	"github.com/ccustine/beastie/types"
	"github.com/gorilla/mux"
	"github.com/r3labs/sse"
	"github.com/rakyll/statik/fs"
	"github.com/rcrowley/go-metrics"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	JSONAPI = "jsonapi"
)

type JsonOutput struct {
	aircraftList []*types.AircraftData
	lock         sync.RWMutex
}

var (
	aircraftList []*types.AircraftData
	lock         sync.RWMutex
	server       *sse.Server
)

func NewJsonOutput() *JsonOutput {
	jsonApi := &JsonOutput{}

	r := mux.NewRouter()
	r.HandleFunc("/aircraft", jsonApi.FeedHandler)
	r.HandleFunc("/metrics", jsonApi.MetricsHandler)

	server = &sse.Server{
		//BufferSize: 1024,
		AutoStream: false,
		AutoReplay: false,
		Streams:    make(map[string]*sse.Stream),
	}
	//sse.New()
	server.CreateStream("aircraft")

	// Create a new Mux and set the handler
	r.HandleFunc("/stream", server.HTTPHandler)

	// This will serve files under http://localhost:8000/static/<filename>
	//r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/dist/"))))
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(statikFS)))

	//r.HandleFunc("/airframes/{icao}", AirframeHandler).Name("article")

	//http.Handle("/", r)

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:8000",
		//WriteTimeout: 15 * time.Second,
		//ReadTimeout:  15 * time.Second,
	}
	go srv.ListenAndServe()

	return jsonApi
}

func (o JsonOutput) UpdateDisplay(knownAircraft []*types.AircraftData) { //*types.AircraftMap) {
	//sortedAircraft := make(AircraftList, 0, aircraftList.Len())

	o.lock.Lock()
	aircraftList = knownAircraft
	o.lock.Unlock()

	//TODO: Don't want to duplicate this everywhere
	var b strings.Builder

	goodRate := metrics.GetOrRegisterMeter("Message Rate (Good)", metrics.DefaultRegistry)
	badRate := metrics.GetOrRegisterMeter("Message Rate (Bad)", metrics.DefaultRegistry)
	modeACCnt := metrics.GetOrRegisterCounter("Message Rate (ModeA/C)", metrics.DefaultRegistry)
	modesShortCnt := metrics.GetOrRegisterCounter("Message Rate (ModeS Short)", metrics.DefaultRegistry)
	modesLongCnt := metrics.GetOrRegisterCounter("Message Rate (ModeS Long)", metrics.DefaultRegistry)

	b.WriteString(fmt.Sprintf(`{"now": %d, "good":%.1f, "bad":%.1f, "modea":%d, "modesshort":%d, "modeslong":%d, "aircraft":[`, time.Now().Unix(), goodRate.Rate1(), badRate.Rate1(), modeACCnt.Count(), modesShortCnt.Count(), modesLongCnt.Count()))
	o.lock.RLock()
	for i, aircraft := range aircraftList {
		acmb, _ := json.Marshal(aircraft)
		b.Write(acmb)
		if i <= len(aircraftList)-2 {
			b.WriteString(",")
		}
	}
	o.lock.RUnlock()

	b.WriteString("]}")

	server.Publish("aircraft", &sse.Event{
		Data: []byte(b.String()),
	})
}

func (o JsonOutput) FeedHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	o.lock.RLock()
	respondWithJSON(w, http.StatusOK, aircraftList)
	o.lock.RUnlock()
}

func (o JsonOutput) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var b strings.Builder

	goodRate := metrics.GetOrRegisterMeter("Message Rate (Good)", metrics.DefaultRegistry)
	badRate := metrics.GetOrRegisterMeter("Message Rate (Bad)", metrics.DefaultRegistry)
	modeACCnt := metrics.GetOrRegisterCounter("Message Rate (ModeA/C)", metrics.DefaultRegistry)
	modesShortCnt := metrics.GetOrRegisterCounter("Message Rate (ModeS Short)", metrics.DefaultRegistry)
	modesLongCnt := metrics.GetOrRegisterCounter("Message Rate (ModeS Long)", metrics.DefaultRegistry)

	o.lock.RLock()
	b.WriteString(fmt.Sprintf(`{"now": %d,
"good":%.1f,
"bad":%.1f,
"modea":%d,
"modesshort":%d,
"modeslong":%d,`, time.Now().Unix(), goodRate.Rate1(), badRate.Rate1(), modeACCnt.Count(), modesShortCnt.Count(), modesLongCnt.Count()))
	o.lock.RUnlock()
	b.WriteString("]}")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(b.String()))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	enableCors(&w)
	var b strings.Builder
	acl := payload.([]*types.AircraftData)

	goodRate := metrics.GetOrRegisterMeter("Message Rate (Good)", metrics.DefaultRegistry)
	badRate := metrics.GetOrRegisterMeter("Message Rate (Bad)", metrics.DefaultRegistry)
	modeACCnt := metrics.GetOrRegisterCounter("Message Rate (ModeA/C)", metrics.DefaultRegistry)
	modesShortCnt := metrics.GetOrRegisterCounter("Message Rate (ModeS Short)", metrics.DefaultRegistry)
	modesLongCnt := metrics.GetOrRegisterCounter("Message Rate (ModeS Long)", metrics.DefaultRegistry)

	b.WriteString(fmt.Sprintf(`{"now": %d,
"total":%d,
"good":%.1f,
"bad":%.1f,
"modea":%d,
"modesshort":%d,
"modeslong":%d,
"aircraft":[`, time.Now().Unix(), len(acl), goodRate.Rate1(), badRate.Rate1(), modeACCnt.Count(), modesShortCnt.Count(), modesLongCnt.Count()))
	for i, aircraft := range acl {
		acmb, _ := json.Marshal(aircraft)
		b.Write(acmb)
		if i <= len(acl)-2 {
			b.WriteString(",")
		}

		//sortedAircraft = append(sortedAircraft, aircraft)
	}

	b.WriteString("]}")

	//response, _ := json.Marshal(sortedAircraft)

	//response := b
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(b.String()))
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

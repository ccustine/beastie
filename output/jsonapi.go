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
	"github.com/ccustine/beastie/types"
	"github.com/gorilla/mux"
	"github.com/rcrowley/go-metrics"
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
}

var (
	aircraftList []*types.AircraftData
	lock sync.RWMutex
)

func NewJsonOutput() *JsonOutput {
	jsonApi := &JsonOutput{}

	r := mux.NewRouter()
	r.HandleFunc("/feed", jsonApi.FeedHandler)
	r.HandleFunc("/metrics", jsonApi.MetricsHandler)

	// This will serve files under http://localhost:8000/static/<filename>
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("/tmp"))))

	//r.HandleFunc("/airframes/{icao}", AirframeHandler).Name("article")

	//http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	//listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8000})
	//if err != nil {
	//	log.Fatal("error creating listener")
	//}

	//go logrus.Error(srv.ListenAndServe())
	//go srv.Serve(listener)
	go srv.ListenAndServe()
	//go log.Fatal(srv.ListenAndServe())
	//go log.Fatal(srv.ListenAndServe())

	return jsonApi
}

func (o JsonOutput) UpdateDisplay(knownAircraft *types.AircraftMap) {
	//sortedAircraft := make(AircraftList, 0, aircraftList.Len())

	lock.Lock()
	aircraftList = knownAircraft.Copy()
	lock.Unlock()

	//var b strings.Builder

	//sortedAircraft := make(AircraftList, 0, aircraftList.Len())

	//for _, aircraft := range aircraftList.Copy() {
	//	sortedAircraft = append(sortedAircraft, aircraft)
	//}

	//sort.Sort(sortedAircraft)

	//for _, aircraft := range aircraftList.Copy() {
	//	evict := time.Since(aircraft.LastPing) > (time.Duration(59) * time.Second)
	//
	//	if evict {
	//		//if o.Beastinfo.Debug {
	//		//	log.Debugf("Evicting %d", aircraft.IcaoAddr)
	//		//}
	//		aircraftList.Delete(aircraft.IcaoAddr)
	//		continue
	//	}
	//
	//	jsonString, _ := json.MarshalIndent(aircraft, "", "\t")
	//	//o.ACLogFile.WriteString(string(jsonString))
	//	//o.Aclog.Info(string(jsonString))
	//}
}

//func (a *App) FeedHandler(w http.ResponseWriter, r *http.Request) {
func (o JsonOutput) FeedHandler(w http.ResponseWriter, r *http.Request) {
	//respondWithJSON(w, http.StatusOK, products)
	lock.RLock()
	respondWithJSON(w, http.StatusOK, aircraftList)
	lock.RUnlock()
}

func (o JsonOutput) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	var b strings.Builder

	goodRate := metrics.GetOrRegisterMeter("Message Rate (Good)", metrics.DefaultRegistry)
	badRate := metrics.GetOrRegisterMeter("Message Rate (Bad)", metrics.DefaultRegistry)
	//ModeACCnt          := metrics.GetOrRegisterCounter("Message Rate (ModeA/C)", metrics.DefaultRegistry)
	//ModesShortCnt      := metrics.GetOrRegisterCounter("Message Rate (ModeS Short)", metrics.DefaultRegistry)
	//ModesLongCnt       := metrics.GetOrRegisterCounter("Message Rate (ModeS Long)", metrics.DefaultRegistry)

	lock.RLock()
	b.WriteString(fmt.Sprintf("{\"now\": %d,\"total\":%d,\"good\":%.1f,\"bad\":%.1f}", time.Now().Unix(), len(aircraftList), goodRate.Rate1(), badRate.Rate1()))
	lock.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(b.String()))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	var b strings.Builder
	acl := payload.([]*types.AircraftData)

	goodRate := metrics.GetOrRegisterMeter("Message Rate (Good)", metrics.DefaultRegistry)
	badRate := metrics.GetOrRegisterMeter("Message Rate (Bad)", metrics.DefaultRegistry)
	//ModeACCnt          := metrics.GetOrRegisterCounter("Message Rate (ModeA/C)", metrics.DefaultRegistry)
	//ModesShortCnt      := metrics.GetOrRegisterCounter("Message Rate (ModeS Short)", metrics.DefaultRegistry)
	//ModesLongCnt       := metrics.GetOrRegisterCounter("Message Rate (ModeS Long)", metrics.DefaultRegistry)

	b.WriteString(fmt.Sprintf("{\"now\": %d,\"total\":%d,\"good\":%.1f,\"bad\":%.1f,\"aircraft\":[", time.Now().Unix(), len(acl), goodRate.Rate1(), badRate.Rate1()))
	for i, aircraft := range acl {
		acmb, _ := json.Marshal(aircraft)
		b.Write(acmb)
		if i <= len(acl) - 2 {
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

func respondWithTestString(w http.ResponseWriter, code int, testString string) {
	response := testString

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(response))
}

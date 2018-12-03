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
	"github.com/ccustine/beastie/types"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	JSONAPI = "jsonapi"
)

type JsonOutput struct {
	knownAircraft types.AircraftMap
}

func NewJsonOutput() *JsonOutput {
	jsonApi := &JsonOutput{}

	r := mux.NewRouter()
	r.HandleFunc("/feed", jsonApi.FeedHandler)

	// This will serve files under http://localhost:8000/static/<filename>
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("/tmp"))))

	//r.HandleFunc("/airframes/{icao}", AirframeHandler).Name("article")

	//http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8000})
	if err != nil {
		log.Fatal("error creating listener")
	}

	go srv.Serve(listener)
	//go log.Fatal(srv.ListenAndServe())
	//go log.Fatal(srv.ListenAndServe())

	return jsonApi
}

func (o JsonOutput) UpdateDisplay(knownAircraft *types.AircraftMap) {
	o.knownAircraft = *knownAircraft
	//var b strings.Builder

	//sortedAircraft := make(AircraftList, 0, knownAircraft.Len())

	//for _, aircraft := range knownAircraft.Copy() {
	//	sortedAircraft = append(sortedAircraft, aircraft)
	//}

	//sort.Sort(sortedAircraft)

	//for _, aircraft := range knownAircraft.Copy() {
	//	evict := time.Since(aircraft.LastPing) > (time.Duration(59) * time.Second)
	//
	//	if evict {
	//		//if o.Beastinfo.Debug {
	//		//	log.Debugf("Evicting %d", aircraft.IcaoAddr)
	//		//}
	//		knownAircraft.Delete(aircraft.IcaoAddr)
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

	respondWithJSON(w, http.StatusOK, o.knownAircraft)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	var b strings.Builder
	acm := payload.(*types.AircraftMap)

	sortedAircraft := make(AircraftList, 0, acm.Len())

	for _, aircraft := range acm.Copy() {
		acmb, _ := json.Marshal(aircraft)
		b.Write(acmb)
		sortedAircraft = append(sortedAircraft, aircraft)
	}


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

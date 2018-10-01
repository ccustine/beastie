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

package main

import (
	"encoding/json"
	"fmt"
	. "github.com/ccustine/beastie/beastie"
	"github.com/ccustine/beastie/modes"
	ver "github.com/ccustine/beastie/version"
	"github.com/gorilla/mux"
	"github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	NewRootCmd().Execute()
}

var (
	VERSION    string
	helpFlag   bool
	debug      bool
	metricflag bool
	beastInfo   = &BeastInfo{}
	adsbSource  = &Source{}
	mlatSource  = &Source{}
)

const LOG_FILE = "/tmp/beastied.log"

//Execute adds all child commands to the root command.
func Execute() {
	//utils.StopOnErr(RootCmd.Execute())
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "beastied",
		Short: "Beastie server daemon",
		Long: `beastied is the main daemon application

Beastie is a Mode-S ES parser and web server.

Read more at https://github.io/ccustine/beastie`,
		Run: func(cmd *cobra.Command, args []string) {
			file, err := os.OpenFile(LOG_FILE, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0755)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			log.SetOutput(file)

			r := mux.NewRouter()
			r.HandleFunc("/feed", FeedHandler)

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

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			go func(){
				//for sig := range c {
				<-c
				if beastInfo.Metrics {
					//spew.Dump(metrics.DefaultRegistry)
					modes.LogOnce(metrics.DefaultRegistry, log.New())
				}
				os.Exit(1)
			}()

			modes.Start(*beastInfo)

		},
		PersistentPreRun: begin,
		PersistentPostRun: end,
	}

	// Persistent == available to sub commands
	rootCmd.PersistentFlags().BoolVarP(&debug, DEBUG, "d", false, "Output debug level logging.")
	rootCmd.PersistentFlags().BoolVarP(&metricflag, METRICS, "m", false, "Output Metrics")
	rootCmd.PersistentFlags().StringVarP(&ConfigFile, CONFIGFILE, "c", "","viper file (default is $HOME/.beastie.yaml)")

	// Must override default help flag to use -h for Host
	rootCmd.PersistentFlags().BoolVarP(&helpFlag, "help", "", false, "Help default flag")
	rootCmd.PersistentFlags().StringVar(&adsbSource.Host, BEAST_HOST, "","Set the BEAST_HOST for the Everyware Cloud API endpoint")
	rootCmd.PersistentFlags().IntVar(&adsbSource.Port, BEAST_PORT, 0, "Beast mode port to connect to")
	rootCmd.PersistentFlags().StringVar(&mlatSource.Host, MLAT_HOST, "", "Set the BEAST_HOST for the Everyware Cloud API endpoint")
	rootCmd.PersistentFlags().IntVar(&mlatSource.Port, MLAT_PORT, 0, "Beast mode port to connect to")
	rootCmd.PersistentFlags().Float64VarP(&beastInfo.Latitude, BASELAT, "", 40.135, "Beast mode port to connect to")
	rootCmd.PersistentFlags().Float64VarP(&beastInfo.Longitude, BASELON, "", -104.997, "Beast mode port to connect to")

	viper.BindPFlag("sources.adsb.host", rootCmd.PersistentFlags().Lookup(BEAST_HOST))
	viper.BindPFlag("sources.adsb.port", rootCmd.PersistentFlags().Lookup(BEAST_PORT))
	viper.BindPFlag("sources.mlat.host", rootCmd.PersistentFlags().Lookup(MLAT_HOST))
	viper.BindPFlag("sources.mlat.port", rootCmd.PersistentFlags().Lookup(MLAT_PORT))
	viper.BindPFlag(BASELAT, rootCmd.PersistentFlags().Lookup(BASELAT))
	viper.BindPFlag(BASELON, rootCmd.PersistentFlags().Lookup(BASELON))

	// for Bash autocomplete
	validSortFlags := []string{"distance", "last", "speed", "alt"}
	rootCmd.PersistentFlags().SetAnnotation("sort", cobra.BashCompOneRequiredFlag, validSortFlags)

	log.SetOutput(os.Stdout)
	cobra.OnInitialize(LoadConfig)

	return rootCmd
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Shows the version number if a final release, or a date for snapshots`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("The version is: ", VERSION)
		ver.PrintVersion()

	},
}

func begin(cmd *cobra.Command, args []string) {
	if debug {
		log.Infoln("Changing to debug logging")
		log.SetLevel(log.DebugLevel)
		beastInfo.Debug = true
	}

	// This might need to be inverted, check for cmd line first
	if !viper.IsSet("sources.adsb") {
		beastInfo.Sources = append(beastInfo.Sources, *adsbSource)
	} else {
		beastInfo.Sources = append(beastInfo.Sources, Source{
			viper.GetString("sources.adsb.host"),
			viper.GetInt("sources.adsb.port")})
	}

	if !viper.IsSet("sources.mlat") {
		beastInfo.Sources = append(beastInfo.Sources, *mlatSource)
	} else {
		beastInfo.Sources = append(beastInfo.Sources, Source{
			viper.GetString("sources.mlat.host"),
			viper.GetInt("sources.mlat.port")})
	}

	//viper.UnmarshalKey("sources.adsb", &adsbSource)
	//viper.UnmarshalKey("sources.mlat", &mlatSource)


	//fmt.Print(viper.AllSettings())
/*
	if host != "" && port != 0 {
		beastInfo.Sources = []modes.Source{{host, port}}
	}
	beastInfo.Latitude = baseLat
	beastInfo.Longitude = baseLon
	beastInfo.Debug = debug
	beastInfo.Metrics = metricflag
*/
	beastInfo.Metrics = metricflag

}

func end(cmd *cobra.Command, args []string) {
}

//func (a *App) FeedHandler(w http.ResponseWriter, r *http.Request) {
func FeedHandler(w http.ResponseWriter, r *http.Request) {
	//respondWithJSON(w, http.StatusOK, products)

	respondWithTestString(w, http.StatusOK, "Test String")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithTestString(w http.ResponseWriter, code int, testString string) {
	response := testString

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(response))
}
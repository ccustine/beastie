package cmd

import (
	"github.com/ccustine/beastie/cmd/registrycmd"
	"github.com/ccustine/beastie/config"
	"os"

	"github.com/ccustine/beastie/cmd/stream"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DEBUG      = "debug"
	METRICS    = "metrics"
	HOST       = "host"
	PORT       = "port"
	BASELAT    = "lat"
	BASELON    = "lon"
	USERNAME   = "user"
	PASSWORD   = "password"
	CONFIGFILE = "cfg"
)

var (
	VERSION   string
	cfgFile   string
	helpFlag  bool
	debug     bool
	metrics   bool
	host      string
	port      int
	baseLat   float64
	baseLon   float64
	beastInfo = &config.BeastInfo{}
)

var cfg = viper.New()

//Execute adds all child commands to the root command.
func Execute() {
	//utils.StopOnErr(RootCmd.Execute())
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "beastie-cli",
		Short: "Command utilities for Kapua",
		Long: `config is the interactive command line

config is the interactive command line for displaying ADS-B data from a Beast / Mode-S radio.

More info at https://github.io/ccustine/config`,
		Run: func(cmd *cobra.Command, args []string) {
			//cmd.Usage()
		},
		PersistentPreRun: begin,
	}

	// Persistent == available to sub commands
	//rootCmd.PersistentFlags().BoolVarP(&debug, DEBUG, "d", false, "Outputs debug level logging.")
	//rootCmd.PersistentFlags().BoolVarP(&metrics, METRICS, "m", false, "Outputs Metrics")
	rootCmd.PersistentFlags().StringVar(&cfgFile, CONFIGFILE, "", "cfg file (default is $HOME/.kapgun.yaml)")

	// Must override default help flag to use -h for Host
	rootCmd.PersistentFlags().BoolVarP(&helpFlag, "help", "", false, "Help default flag")
	rootCmd.PersistentFlags().StringVarP(&host, HOST, "h", "rpi3-1-wifi.home.custine.com", "Set the HOST for the Everyware Cloud API endpoint")
	rootCmd.PersistentFlags().IntVarP(&port, PORT, "p", 30005, "Beast mode port to connect to")
	rootCmd.PersistentFlags().Float64VarP(&baseLat, BASELAT, "", 40.135, "Beast mode port to connect to")
	rootCmd.PersistentFlags().Float64VarP(&baseLon, BASELON, "", -104.997, "Beast mode port to connect to")

	cfg.BindPFlag("sources."+HOST, rootCmd.PersistentFlags().Lookup(HOST))
	cfg.BindPFlag("sources."+PORT, rootCmd.PersistentFlags().Lookup(PORT))
	cfg.BindPFlag(BASELAT, rootCmd.PersistentFlags().Lookup(BASELAT))
	cfg.BindPFlag(BASELON, rootCmd.PersistentFlags().Lookup(BASELON))

	// for Bash autocomplete
	validSortFlags := []string{"distance", "last", "speed", "alt"}
	rootCmd.PersistentFlags().SetAnnotation("sort", cobra.BashCompOneRequiredFlag, validSortFlags)

	rootCmd.AddCommand(stream.NewRootCmd(beastInfo))
	rootCmd.AddCommand(registrycmd.NewImportCmd(beastInfo))
	rootCmd.AddCommand(registrycmd.NewDownloadCmd(beastInfo))
	rootCmd.AddCommand(registrycmd.NewListCmd(beastInfo))
	rootCmd.AddCommand(registrycmd.NewFindCmd(beastInfo))
	rootCmd.AddCommand(NewVersionCmd())

	log.SetOutput(os.Stdout)
	cobra.OnInitialize(config.LoadConfig)

	return rootCmd
}

func begin(cmd *cobra.Command, args []string) {
	if debug {
		log.Infoln("Changing to debug logging")
		log.SetLevel(log.DebugLevel)
	}
	//beastInfo.Host = host
	//beastInfo.Port = port
	beastInfo.Latitude = baseLat
	beastInfo.Longitude = baseLon
	beastInfo.Debug = debug
	beastInfo.Metrics = metrics
}

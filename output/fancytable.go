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
	"bytes"
	"fmt"
	"github.com/ccustine/beastie/config"
	registration "github.com/ccustine/beastie/db"
	"github.com/ccustine/beastie/output/termui"
	"github.com/ccustine/beastie/registry"
	"github.com/ccustine/beastie/types"
	"github.com/dgraph-io/badger"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/kellydunn/golang-geo"
	"github.com/mattn/go-isatty"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"image"
	"math"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	helpVisible bool
	renderLock  sync.Mutex
)

const (
	FANCYTABLE = "fancytable"
	UP         = "▲"
	DOWN       = "▼"
)

type FancyTable struct {
	*termui.Table
	h          *HelpMenu
	aircraft   []types.AircraftData
	act        *termui.Table
	g          *ui.Grid
	msgRate    *widgets.Plot
	Beastinfo  *config.BeastInfo
	i          *widgets.Paragraph
	sortMethod string
	sortAsc    bool
	acinfo     *widgets.Paragraph
	db         *badger.DB
	isClosing  bool
	group      *sync.WaitGroup
	done       chan<- interface{}
}

func render(drawable ...ui.Drawable) {
	renderLock.Lock()
	ui.Render(drawable...)
	renderLock.Unlock()
}

func (o *FancyTable) startSelfUpdateTicker() {
	go func() {
		ticker := time.NewTicker(5000 * time.Millisecond) //TODO: Make this adjustable and separate tickers per output
		for {
			select {
			case <-ticker.C:
				GoodRate := metrics.GetOrRegisterMeter("Message Rate (Good)", metrics.DefaultRegistry)
				//BadRate := metrics.GetOrRegisterMeter("Message Rate (Bad)", metrics.DefaultRegistry)
				goodRate := math.Round(GoodRate.Rate1())
				//badRate := BadRate.Rate1()

				if goodRate > o.msgRate.MaxVal {
					o.msgRate.MaxVal = math.Round(goodRate/100)*100 + 100
				}

				if o.msgRate.Data[0][0] == 100 {
					o.msgRate.Data[0][0] = goodRate
					o.msgRate.Data[0][1] = goodRate
				} else if o.msgRate.Dx() != 0 && len(o.msgRate.Data[0]) > o.msgRate.Dx()-8 {
					o.msgRate.Data[0] = append(o.msgRate.Data[0][1:], goodRate)
				} else {
					o.msgRate.Data[0] = append(o.msgRate.Data[0], goodRate)
				}

				/*				if o.msgRate.Dx() != 0 && len(o.msgRate.Data[1]) > o.msgRate.Dx()-8 {
									o.msgRate.Data[1] = append(o.msgRate.Data[1][1:], badRate)
								} else {
									o.msgRate.Data[1] = append(o.msgRate.Data[1], badRate)
								}
				*/
				if !helpVisible {
					renderLock.Lock()
					ui.Render(o.msgRate)
					renderLock.Unlock()

				}
			}
		}
	}()
}

func (o *FancyTable) startPollUi() {
	go func() {
		sigTerm := make(chan os.Signal, 2)
		signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

		uiEvents := ui.PollEvents()

		previousKey := ""

		for {
			select {
			case <-sigTerm:
				//o.isClosing = true
				o.Close()
				/*		case <-drawTicker:
						if !helpVisible {
							render(grid)
						}
				*/
			case e := <-uiEvents:
				switch e.ID {
				case "q", "<C-c>":
					o.Close()
				case "?":
					helpVisible = !helpVisible
				case "<Resize>":
					payload := e.Payload.(ui.Resize)
					//termWidth, termHeight := payload.Width, payload.Height
					//if statusbar {
					//	o.g.SetRect(0, 0, termWidth, termHeight-1)
					//	bar.SetRect(0, termHeight-1, termWidth, termHeight)
					//} else {
					o.g.SetRect(0, 0, payload.Width, payload.Height)
					//}
					//help.Resize(payload.Width, payload.Height)
					ui.Clear()

				}
				if helpVisible {
					switch e.ID {
					case "?":
						ui.Clear()
						ui.Render(o.h)
					case "<Escape>":
						helpVisible = false
						render(o.g)
					case "<Resize>":
						ui.Render(o.h)
					}
				} else {

					switch e.ID {
					case "?":
						render(o.g)
						/*								case "h":
															graphHorizontalScale += graphHorizontalScaleDelta
															cpu.HorizontalScale = graphHorizontalScale
															mem.HorizontalScale = graphHorizontalScale
															render(cpu, mem)
														case "acft":
															if graphHorizontalScale > graphHorizontalScaleDelta {
																graphHorizontalScale -= graphHorizontalScaleDelta
																cpu.HorizontalScale = graphHorizontalScale
																mem.HorizontalScale = graphHorizontalScale
																render(cpu, mem)
															}
						*/
					case "<Resize>":
						render(o.g)
						/*									if statusbar {
																render(bar)
															}
						*/
					case "<MouseLeft>":
						payload := e.Payload.(ui.Mouse)
						o.HandleClick(payload.X, payload.Y)
						render(o)
					case "k", "<Up>", "<MouseWheelUp>":
						o.ScrollUp()
						render(o)
					case "j", "<Down>", "<MouseWheelDown>":
						o.ScrollDown()
						render(o)
					case "<Home>":
						o.ScrollTop()
						render(o)
					case "g":
						if previousKey == "g" {
							o.ScrollTop()
							render(o)
						}
					case "G", "<End>":
						o.ScrollBottom()
						render(o)
					case "<C-d>":
						o.ScrollHalfPageDown()
						render(o)
					case "<C-u>":
						o.ScrollHalfPageUp()
						render(o)
					case "<C-f>":
						o.ScrollPageDown()
						render(o)
					case "<C-b>":
						o.ScrollUp()
						render(o)
					case "d":
						if previousKey == "d" {
							o.Select()
						}
					case "<Enter>":
						o.Select()
					case "<Tab>":
						o.Tab()
						render(o)
					case "r", "s":
						o.ChangeSort(e)
						render(o)
					}
				}

				if previousKey == e.ID {
					previousKey = ""
				} else {
					previousKey = e.ID
				}
			}
		}
	}()
}

func NewFancyTableOutput(info *config.BeastInfo, group *sync.WaitGroup, done chan<- interface{}) *FancyTable {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		logrus.Infof("This is a terminal, enabling UI output")
	} else if isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		fmt.Println("Is Cygwin/MSYS2 Terminal")
	} else {
		logrus.Error("This is not a terminal, disabling UI output")
		return nil
	}

	if err := ui.Init(); err != nil {
		logrus.Fatalf("failed to initialize termui: %v", err)
	}
	h := NewHelpMenu()

	act := termui.NewTable()
	//act.Title = " Aircraft List "

	//act.ColWidths = []int{
	//	3, 6, 8, 4, 15, 5, 4, 4, 3, 3, 5, 4,
	//}

	act.Header = []string{"#", "ICAO", "Call", "Squawk", "Lat/Lon", "Alt", "Rate", "Speed", "Hdg", "Rng", "Last"}

	//act.Rows = make([][]string, 2)
	//act.Rows[0] = []string{"#", "ICAO", "Call", "Squawk", "Lat/Lon", "Alt", "Rate", "Speed", "Hdg", "Rng", "Last"}

	act.ColResizer = func() {
		act.ColWidths = []int{
			4, 7, 9, 7, 18, 8, 6, 6, 4, 6, 6,
		}
	}

	infoPar := widgets.NewParagraph()

	acInfo := widgets.NewParagraph()

	msgRate := widgets.NewPlot()
	msgRate.Title = "Msg Rate"
	msgRate.Data = make([][]float64, 1)
	msgRate.Data[0] = []float64{100,100}
	//msgRate.Data[1] = []float64{1, 1}
	msgRate.DataLabels = []string{"Good/s", "Bad/s"}
	msgRate.ShowAxes = true
	//msgRate.HorizontalScale = 2
	msgRate.LineColors = []ui.Color{ui.ColorGreen, ui.ColorRed}
	//msgRate.DrawDirection = widgets.DrawLeft
	//msgRate.LineColor = ui.ColorGreen
	msgRate.MaxVal = 100

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		/*		ui.NewRow(1.0/5,
					ui.NewCol(1.0/3*2, msgRate),
					ui.NewCol(1.0/3, infoPar),
				),
		*/ui.NewRow(1.0,
			ui.NewCol(1.0/3*2, act),
			ui.NewCol(1.0/3,
				ui.NewRow(1.0/8, msgRate),
				ui.NewRow(1.0/8, infoPar),
				ui.NewRow(1.0/8*6, acInfo),
			),

		),
	)

	h.Resize(termWidth, termHeight)

	here = geo.NewPoint(info.Latitude, info.Longitude)

	opts := badger.DefaultOptions
	opts.Logger = logrus.StandardLogger()
	opts.Dir = ".data"
	opts.ValueDir = ".data"
	db, err := badger.Open(opts)
	checkErr(err)

	group.Add(1)
	table := &FancyTable{Beastinfo: info, act: act, done: done, group: group, sortMethod: "r", sortAsc: true, db: db, acinfo: acInfo, i: infoPar, h: h, g: grid, Table: act, msgRate: msgRate,}
	table.CursorColor = ui.ColorCyan
	table.ShowCursor = true
	table.UniqueCol = 1

	//renderLock.Lock()
	ui.Render(grid)
	//renderLock.Unlock()

	table.startPollUi()
	table.startSelfUpdateTicker()

	return table
}

func (o *FancyTable) Close() {
	o.isClosing = true
	o.done <- true
	o.db.Close()
	ui.Close()
	o.group.Done()
}

func (o *FancyTable) UpdateDisplay(knownAircraft []*types.AircraftData) { //*types.AircraftMap) {
	aircraftList := make([]types.AircraftData, 0, len(knownAircraft)) //.Len())

	for _, aircraft := range knownAircraft { //.Copy() {
		aircraftList = append(aircraftList, *aircraft)
	}

	o.aircraft = aircraftList
	o.Sort()

	if !helpVisible {
		renderLock.Lock()
		ui.Render(o)
		renderLock.Unlock()
	}

	GoodRate := metrics.GetOrRegisterMeter("Message Rate (Good)", metrics.DefaultRegistry)
	BadRate := metrics.GetOrRegisterMeter("Message Rate (Bad)", metrics.DefaultRegistry)
	RtlGoodRate := metrics.GetOrRegisterMeter("Message Rate (RTL Good)", metrics.DefaultRegistry)
	RtlBadRate := metrics.GetOrRegisterMeter("Message Rate (RTL Bad)", metrics.DefaultRegistry)

	goodRate := GoodRate.Rate1()
	badRate := BadRate.Rate1()

	//if goodRate > o.msgRate.MaxVal {
	//	o.msgRate.MaxVal = (goodRate / 100 * 100) + 100
	//}
	//
	//if o.msgRate.Dx() != 0 && len(o.msgRate.Data[0]) > o.msgRate.Dx()-8 {
	//	o.msgRate.Data[0] = append(o.msgRate.Data[0][1:], goodRate)
	//} else {
	//	o.msgRate.Data[0] = append(o.msgRate.Data[0], goodRate)
	//}
	//
	//if o.msgRate.Dx() != 0 && len(o.msgRate.Data[1]) > o.msgRate.Dx()-8 {
	//	o.msgRate.Data[1] = append(o.msgRate.Data[1][1:], badRate)
	//} else {
	//	o.msgRate.Data[1] = append(o.msgRate.Data[1], badRate)
	//}

	o.msgRate.Title = fmt.Sprintf(" %.1f/s Good : %.1f/s Bad ", goodRate, badRate)

	ModeACCnt := metrics.GetOrRegisterCounter("Message Rate (ModeA/C)", metrics.DefaultRegistry)
	ModesShortCnt := metrics.GetOrRegisterCounter("Message Rate (ModeS Short)", metrics.DefaultRegistry)
	ModesLongCnt := metrics.GetOrRegisterCounter("Message Rate (ModeS Long)", metrics.DefaultRegistry)

	o.i.Text = fmt.Sprintf("Message Rate [(Good)](fg:green): %.1f/s\nMessage Rate [(Bad)](fg:red) : %.1f/s\nMessage Rate [(RTL Good)](fg:green) : %.1f/s\nMessage Rate [(RTL Bad)](fg:red) : %.1f/s\n", goodRate, badRate, RtlGoodRate.Rate1(), RtlBadRate.Rate1()) +
		fmt.Sprintf("Message Count - Mode A/C:    %d\nMessage Count - ModeS Short: %d\nMessage Count - ModeS Long:  %d", ModeACCnt.Count(), ModesShortCnt.Count(), ModesLongCnt.Count())

	if !helpVisible {
		renderLock.Lock()
		ui.Render(o.msgRate)
		ui.Render(o.i)
		renderLock.Unlock()

	}

}

// Sort sorts either the grouped or ungrouped []Process based on the sortMethod.
// Called with every update, when the sort method is changed, and when processes are grouped and ungrouped.
func (o *FancyTable) Sort() {
	aircraftData := o.aircraft
	o.Header = []string{"#", "ICAO", "Call", "Squawk", "Lat/Lon", "Alt", "Rate", "Speed", "Hdg", "Rng", "Last"}

	switch o.sortMethod {
	case "s":
		if o.sortAsc {
			sort.Sort(AircraftBySpeed(aircraftData))
			o.Header[7] += UP
		} else {
			sort.Sort(sort.Reverse(AircraftBySpeed(aircraftData)))
			o.Header[7] += DOWN
		}
	case "r":
		if o.sortAsc {
			sort.Sort(AircraftByRange(aircraftData))
			o.Header[9] += UP
		} else {
			sort.Sort(sort.Reverse(AircraftByRange(aircraftData)))
			o.Header[9] += DOWN
		}
	}

	//o.Rows, o.ColStyles = o.FieldsToStrings(aircraftData)
	o.Rows, _ = o.FieldsToStrings(aircraftData)
}

func (o *FancyTable) ChangeSort(e ui.Event) {
	if o.sortMethod != e.ID {
		o.sortMethod = e.ID
		o.ScrollTop()
		o.Sort()
	} else {
		o.sortAsc = !o.sortAsc
		o.ScrollTop()
		o.Sort()
	}
}

func (o *FancyTable) Tab() {
	/*	o.group = !o.group
		if o.group {
			o.UniqueCol = 1
		} else {
			o.UniqueCol = 0
		}
		o.Sort()
		o.Top()
	*/
}

// FieldsToStrings converts a []Process to a [][]string
func (o *FancyTable) FieldsToStrings(sortedAircraft []types.AircraftData) ([][]string, [][]ui.Style) {
	var rows [][]string //:= make([][]string), len(sortedAircraft))

	styles := make([][]ui.Style, len(sortedAircraft))
	for i := range styles {
		styles[i] = make([]ui.Style, 11)
	}

	index := 0
	for _, aircraft := range sortedAircraft {
		aircraftHasLocation := aircraft.Latitude != math.MaxFloat64 &&
			aircraft.Longitude != math.MaxFloat64
		// This hides ac with no pos from the display
		if !aircraftHasLocation {
			continue
		}
		index += 1
		aircraftHasAltitude := aircraft.Altitude != math.MaxInt32

		var sLatLon string
		var sAlt string

		isMlat := ""
		if aircraft.Mlat {
			isMlat = "*"
		}

		if aircraftHasLocation {
			sLatLon = fmt.Sprintf("%3.3f, %3.3f%s", aircraft.Latitude, aircraft.Longitude, isMlat)
		} else {
			sLatLon = "---.------,---.------"
		}

		if aircraftHasAltitude && aircraft.Surface == false {
			// TODO: This is noisy, need to figure out how to smooth and watch trending
			var vrs string
			if aircraft.VertRate >= 250 {
				switch aircraft.VertRateSign {
				case 0:
					vrs = "➚"
				case 1:
					vrs = "➘"
				default:
					vrs = ""
				}
			} else {
				vrs = ""
			}

			sAlt = fmt.Sprintf("%d %s", aircraft.Altitude, vrs)
		} else if aircraft.Surface == true {
			sAlt = "Grnd"
		} else {
			sAlt = "-----"
		}

		acpos := geo.NewPoint(aircraft.Latitude, aircraft.Longitude)
		homepos := geo.NewPoint(o.Beastinfo.Latitude, o.Beastinfo.Longitude)
		dist := homepos.GreatCircleDistance(acpos)

		distance := dist * 0.539957 // nm //0.621371 - statue mile

		tPing := time.Since(aircraft.LastPing)

		/*
			//var rowcolor color.Color
			stale := tPing > (time.Duration(10) * time.Second)
			pendingEvict := tPing > (time.Duration(35) * time.Second)
			if !stale && !pendingEvict {
				o.act.Ro = theme[0]
			} else if stale && !pendingEvict {
				rowcolor = theme[1]
			} else if pendingEvict {
				rowcolor = theme[2]
			}
		*/
		tPos := time.Since(aircraft.LastPos)
		posstale := tPos > (time.Duration(10) * time.Second)
		pospendingEvict := tPos > (time.Duration(20) * time.Second)

		if !posstale && !pospendingEvict {
			styles[index][4] = ui.NewStyle(ui.ColorWhite)
		} else if posstale && !pospendingEvict {
			styles[index][4] = ui.NewStyle(ui.ColorYellow)
		} else if pospendingEvict {
			styles[index][4] = ui.NewStyle(ui.ColorRed)
		} else {
			styles[index][4] = ui.NewStyle(ui.ColorWhite)
		}

		var vertRate string
		if aircraft.VertRate >= 250 {
			vertRate = fmt.Sprintf("%d", aircraft.VertRate)
		} else {
			vertRate = ""
		}

		var squawk string
		if aircraft.Squawk > 0 {
			squawk = fmt.Sprintf("%04x", aircraft.Squawk)
		} else {
			squawk = ""
		}

		mil := ""
		if aircraft.Military {
			mil = "*"
		}

		rows = append(rows,
			[]string{
				fmt.Sprintf("%d", index),
				fmt.Sprintf("%06x%s", aircraft.IcaoAddr, mil),
				aircraft.Callsign,
				squawk, //"[test](fg:red)",
				sLatLon,
				sAlt,
				vertRate,
				fmt.Sprintf("%d", aircraft.Speed),
				fmt.Sprintf("%d", aircraft.Heading),
				fmt.Sprintf("%3.1f", distance),
				fmt.Sprintf("%2d", uint8(tPing.Seconds())),
			})
	}

	return rows, styles

}

// Select looks up the aircraft info in the registry DB and displays it in the proper cell
func (o *FancyTable) Select() {
	o.SelectedItem = ""
	icaotmp, _ := strconv.ParseUint(o.Rows[o.SelectedRow][o.UniqueCol], 16, 32)
	icao := uint32(icaotmp)
	country := registration.IcaoToCountry(icao)
	o.acinfo.Text = fmt.Sprintf("Aircraft Info: %s", o.Rows[o.SelectedRow][o.UniqueCol])

	//start := time.Now()

	mkey := bytes.NewBufferString(registration.AIRCRAFT_PREFIX)
	mkey.Write(bytes.ToUpper([]byte(o.Rows[o.SelectedRow][o.UniqueCol])))

	errdb := o.db.View(func(txn *badger.Txn) error {
		mitem, err := txn.Get(mkey.Bytes())
		if err != nil {
			//o.acinfo.Text = fmt.Sprintf("Error in lookup: %s\nKey: %s", err, mkey.String())
			return err
		}

		master := &registry.Master{}
		err = mitem.Value(func(val []byte) error {
			err = registration.DecodeMsgPack(val, master)
			return err
		})

		rkey := bytes.NewBufferString(registration.REGISTRATION_PREFIX)
		rkey.Write([]byte(master.MFR_MDL_CODE))
		ritem, err := txn.Get(rkey.Bytes())
		if err != nil {
			o.acinfo.Text = fmt.Sprintf("Error in ref lookup: %s\nKey: %s", err, rkey.String())
			return nil
		}

		ref := &registry.AircraftRef{}
		err = ritem.Value(func(val []byte) error {
			err = registration.DecodeMsgPack(val, ref)
			return err
		})

		o.acinfo.Title = fmt.Sprintf(" Aircraft: %s - Tail Number: %s ", master.MODE_S_CODE_HEX, master.N_NUMBER)
		o.acinfo.Text = fmt.Sprintf("Aircraft Make: [%s](fg:cyan)\nModel: [%s](fg:cyan)\nYear: [%s](fg:cyan)\n"+
			"Owner: [%s](fg:cyan)\nCountry: [%s](fg:cyan)\nFA: https://flightaware.com/photos/aircraft/N%s", ref.MFR, ref.MODEL, master.YEAR_MFR, master.NAME, country, master.N_NUMBER)

		return err
	})

	if errdb != nil {
		o.acinfo.Text = fmt.Sprintf("Error in lookup: %s\nKey: %s\nCountry: %s", errdb, mkey.String(), country)
	}

	//renderLock.Lock()
	render(o.acinfo)
	//renderLock.Unlock()

}

const KEYBINDS = `
Quit: q or <C-c>

Process navigation
  - k and <Up>: up
  - j and <Down>: down
  - <C-u>: half page up
  - <C-d>: half page down
  - <C-b>: full page up
  - <C-f>: full page down
  - gg and <Home>: jump to top
  - G and <End>: jump to bottom

Process actions:
  - <Tab>: toggle process grouping
  - dd: kill selected process or group of processes

Process sorting
  - c: CPU
  - m: Mem
  - p: PID

CPU and Mem graph scaling:
  - h: scale in
  - l: scale out
`

type HelpMenu struct {
	ui.Block
}

func NewHelpMenu() *HelpMenu {
	return &HelpMenu{
		Block: *ui.NewBlock(),
	}
}

func (self *HelpMenu) Resize(termWidth, termHeight int) {
	textWidth := 53
	textHeight := 22
	x := (termWidth - textWidth) / 2
	y := (termHeight - textHeight) / 2

	self.Block.SetRect(x, y, textWidth+x, textHeight+y)
}

func (self *HelpMenu) Draw(buf *ui.Buffer) {
	self.Block.Draw(buf)

	for y, line := range strings.Split(KEYBINDS, "\n") {
		for x, char := range line {
			buf.SetCell(
				ui.NewCell(char, ui.NewStyle(7)),
				image.Pt(self.Inner.Min.X+x, self.Inner.Min.Y+y-1),
			)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////////
//                              []Process Sorting                              //
/////////////////////////////////////////////////////////////////////////////////

type AircraftByRange []types.AircraftData

// Len implements Sort interface
func (P AircraftByRange) Len() int {
	return len(P)
}

// Swap implements Sort interface
func (P AircraftByRange) Swap(i, j int) {
	P[i], P[j] = P[j], P[i]
}

// Less implements Sort interface
func (P AircraftByRange) Less(i, j int) bool {
	return P[i].Range < P[j].Range
}

type AircraftBySpeed []types.AircraftData

// Len implements Sort interface
func (P AircraftBySpeed) Len() int {
	return len(P)
}

// Swap implements Sort interface
func (P AircraftBySpeed) Swap(i, j int) {
	P[i], P[j] = P[j], P[i]
}

// Less implements Sort interface
func (P AircraftBySpeed) Less(i, j int) bool {
	return P[i].Speed < P[j].Speed
}

/*type ProcessByPID []Process

// Len implements Sort interface
func (P ProcessByPID) Len() int {
	return len(P)
}

// Swap implements Sort interface
func (P ProcessByPID) Swap(i, j int) {
	P[i], P[j] = P[j], P[i]
}

// Less implements Sort interface
func (P ProcessByPID) Less(i, j int) bool {
	return P[i].PID < P[j].PID
}

type ProcessByMem []Process

// Len implements Sort interface
func (P ProcessByMem) Len() int {
	return len(P)
}

// Swap implements Sort interface
func (P ProcessByMem) Swap(i, j int) {
	P[i], P[j] = P[j], P[i]
}

// Less implements Sort interface
func (P ProcessByMem) Less(i, j int) bool {
	return P[i].Mem < P[j].Mem
}*/

/*func getRegistration(icao string) AircraftData {
	start := time.Now()


	errdb := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(AIRCRAFT_PREFIX)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			master := &registry.Master{}
			item := it.Item()
			//k := item.Key()
			err := item.Value(func(v []byte) error {
				err := decodeMsgPack(v, master)
				//fmt.Printf("key=%s, value=%s\n", k, v)
				return err
			})
			logrus.Infof("Decoded Aircraft: %+v", master)
			if err != nil {
				return err
			}
		}
		return nil
	})

	checkErr(errdb)

	elapsed := time.Since(start)
	fmt.Printf("%02d:%02d:%02d Elapsed...\n", elapsed/time.Hour, elapsed/time.Minute, elapsed/time.Second)
}
*/

func checkErr(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}

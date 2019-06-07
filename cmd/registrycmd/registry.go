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

package registrycmd

import (
	"bytes"
	"encoding/csv"
	"fmt"
	registration "github.com/ccustine/beastie/db"
	"github.com/ccustine/beastie/registry"
	"github.com/dgraph-io/badger"
	"github.com/jszwec/csvutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	MASTER_IN           = "min"
	ACREF_IN            = "ain"
	DBDIR               = "dbdir"
	IMPORT_DIR          = ".import"
	DRY_RUN             = "dry run"
)

var (
	minfile     string
	ainfile     string
	dbdir       string
	import_path string
	dry_run     bool
	year        string
	p           *mpb.Progress
)

type ImportDescriptor struct {
	File   string
	Prefix string
	Target interface{}
}

type keyConverter func(interface{}) string

func NewImportRegistryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: importFunc,
		PreRun: setupImport,
	}
	cmd.Flags().StringVarP(&minfile, MASTER_IN, "m", "", "csv file to import from")
	cmd.Flags().StringVarP(&ainfile, ACREF_IN, "a", "", "csv file to import from")
	cmd.Flags().StringVarP(&dbdir, DBDIR, "d", ".data", "DB directory")
	cmd.Flags().BoolVarP(&dry_run, DRY_RUN, "n", false, "Dry run (don't save to BoltDB")
	return cmd
}

func NewDownloadRegistryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: downloadFunc,
		PreRun: setupImport,
	}
	cmd.Flags().StringVarP(&year, registry.YEAR, "y", "", "Registration DB Year to download")
	cmd.Flags().StringVarP(&dbdir, DBDIR, "d", ".data", "DB directory")

	return cmd
}

func NewListRegistryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: listFunc,
	}
	cmd.Flags().StringVarP(&minfile, MASTER_IN, "m", "", "csv file to import from")
	cmd.Flags().StringVarP(&ainfile, ACREF_IN, "a", "", "csv file to import from")
	cmd.Flags().StringVarP(&dbdir, DBDIR, "d", ".data", "DB directory")
	cmd.Flags().BoolVarP(&dry_run, DRY_RUN, "n", false, "Dry run (don't save to BoltDB")
	return cmd
}

func NewFindRegistryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: listFunc,
	}
	cmd.Flags().StringVarP(&minfile, MASTER_IN, "m", "", "csv file to import from")
	cmd.Flags().StringVarP(&ainfile, ACREF_IN, "a", "", "csv file to import from")
	cmd.Flags().StringVarP(&dbdir, DBDIR, "d", ".data", "DB directory")
	cmd.Flags().BoolVarP(&dry_run, DRY_RUN, "n", false, "Dry run (don't save to BoltDB")
	return cmd
}

func setupImport(cmd *cobra.Command, args []string) {
	import_path = dbdir + string(filepath.Separator) + IMPORT_DIR + string(filepath.Separator)
	if _, err := os.Stat(import_path); os.IsNotExist(err) {
		err = os.MkdirAll(import_path, 0755)
		checkErr(err)
	}
}

func listFunc(cmd *cobra.Command, args []string) {
	start := time.Now()

	opts := badger.DefaultOptions
	opts.Dir = dbdir
	opts.ValueDir = dbdir
	db, err := badger.Open(opts)
	defer db.Close()
	checkErr(err)

	errdb := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(registration.AIRCRAFT_PREFIX)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			master := &registry.Master{}
			item := it.Item()
			//k := item.Key()
			err := item.Value(func(v []byte) error {
				err := registration.DecodeMsgPack(v, master)
				//fmt.Printf("key=%s, value=%s\n", k, v)
				return err
			})
			logrus.Infof("Decoded Master key: %s %+v", string(item.Key()), master)
			if err != nil {
				return err
			}
		}

		it = txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix = []byte(registration.REGISTRATION_PREFIX)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			ref := &registry.AircraftRef{}
			item := it.Item()
			//k := item.Key()
			err := item.Value(func(v []byte) error {
				err := registration.DecodeMsgPack(v, ref)
				//fmt.Printf("key=%s, value=%s\n", k, v)
				return err
			})
			logrus.Infof("Decoded Aircraft key: %s %+v", string(item.Key()), ref)
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

func findFunc(cmd *cobra.Command, args []string) {
	start := time.Now()

	opts := badger.DefaultOptions
	opts.Dir = dbdir
	opts.ValueDir = dbdir
	db, err := badger.Open(opts)
	defer db.Close()
	checkErr(err)

	errdb := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(registration.AIRCRAFT_PREFIX)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			master := &registry.Master{}
			item := it.Item()
			//k := item.Key()
			err := item.Value(func(v []byte) error {
				err := registration.DecodeMsgPack(v, master)
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

func importFunc(cmd *cobra.Command, args []string) {
	start := time.Now()
	doneWg := new(sync.WaitGroup)

	opts := badger.DefaultOptions
	opts.Dir = dbdir
	opts.ValueDir = dbdir
	db, err := badger.Open(opts)
	defer db.Close()

	checkErr(err)

	p = mpb.New(mpb.WithWidth(64), mpb.WithWaitGroup(doneWg), mpb.WithRefreshRate(250*time.Millisecond))

	doneWg.Add(2)
	go importFile(doneWg, db, import_path + minfile, registration.AIRCRAFT_PREFIX, &registry.Master{}, registry.USMasterKey)
	go importFile(doneWg, db, import_path + ainfile, registration.REGISTRATION_PREFIX, &registry.AircraftRef{}, registry.USRegistrationKey)

	p.Wait()
	elapsed := time.Since(start)
	fmt.Printf("%02d:%02d:%02d Elapsed...\n", elapsed/time.Hour, elapsed/time.Minute, elapsed/time.Second)
}

func downloadFunc(cmd *cobra.Command, args []string) {
	doneWg := new(sync.WaitGroup)
	prg := mpb.New(mpb.WithWidth(64), mpb.WithRefreshRate(250*time.Millisecond))

	doneWg.Add(3)
	go registry.DownloadFile(doneWg, import_path + registry.US_REGISTRATION_FILENAME, registry.CreateReleasableUrl(year), prg)
	go registry.DownloadFile(doneWg, import_path + registry.CA_REGISTRATION_FILENAME, registry.CA_REGISTRATION_URL, prg)
	go registry.DownloadFile(doneWg, import_path + registry.AU_REGISTRATION_FILENAME, registry.AU_REGISTRATION_URL, prg)
	//checkErr(err)

	doneWg.Wait()
	fmt.Println("Downloads Finished")

}

func importFile(wg *sync.WaitGroup, db *badger.DB, infile string, prefix string, value interface{}, getKey keyConverter) {
	defer wg.Done()

	if prefix == registration.AIRCRAFT_PREFIX || prefix == registration.REGISTRATION_PREFIX {
		if infile != "" {
			// Modify and write to temp file
			// Open original file
			var file, err = os.OpenFile(infile, os.O_RDWR, 0644)
			if isError(err) { return }
			defer file.Close()

			// read file, line by line
			var text = make([]byte, 1024)
			for {
				line, err = file.Read(text)

				line = "," + line

				// break if finally arrived at end of file
				if err == io.EOF {
					break
				}

				// break if error occured
				if err != nil && err != io.EOF {
					isError(err)
					break
				}
			}

			fmt.Println("==> done reading from file")
			fmt.Println(string(text))



			// open file using READ & WRITE permission
			var file, err = os.OpenFile(path, os.O_RDWR, 0644)
			if isError(err) { return }
			defer file.Close()

			// write some text line-by-line to file
			_, err = file.WriteString("halo\n")
			if isError(err) { return }
			_, err = file.WriteString("mari belajar golang\n")
			if isError(err) { return }

			// save changes
			err = file.Sync()
			if isError(err) { return }

			fmt.Println("==> done writing to file")
			infile = "TEMP FILE NAME"

		}
	}

	if infile != "" {
		f, err := os.Open(infile)
		checkErr(err)

		lc, _ := lineCounter(f)
		f.Seek(0, 0)
		task := filepath.Base(infile)
		job := " importing "
		b := p.AddBar(int64(lc),
			mpb.PrependDecorators(
				decor.Name(task, decor.WC{W: len(task) + 1, C: decor.DSyncWidthR}),
				decor.Name(job, decor.WC{W: len(job) + 1, C: decor.DSyncWidth}),
				decor.OnComplete(decor.CountersNoUnit("%d / %d", decor.WCSyncWidthR), fmt.Sprintf("%d Done!", lc)),
			),
			mpb.AppendDecorators(decor.Percentage(decor.WC{W: 5})),
		)

		csvReader := csv.NewReader(f)
		csvReader.LazyQuotes = true
		dec, err := csvutil.NewDecoder(csvReader)
		checkErr(err)

		dec.Map = registry.TrimSpace

		wb := db.NewWriteBatch()
		defer wb.Cancel()

		var i int32
		for {
			i++
			b.Increment()
			if err := dec.Decode(value); err == io.EOF {
				break
			} else if err != nil {
				logrus.Fatal(err)
			}
			if !dry_run {
				v, _ := registration.EncodeMsgPack(value)
				/*				err := db.Update(func(txn *badger.Txn) error {
									err := txn.Set([]byte(prefix+getKey(value)), v.Bytes())
									return err
								})
				*/
				err := wb.Set([]byte(prefix+getKey(value)), v.Bytes(), 0) // Will create txns as needed.
				checkErr(err)
			}
		}
		err = wb.Flush()
		checkErr(err)
	}
}

func checkErr(err error) {
	if err != nil {
		logrus.Fatal(err.Error())
	}
}

func isError(err error) bool {
	if err != nil {
		logrus.Error(err.Error())
	}

	return (err != nil)
}

func lineCounter(r io.Reader) (int, error) {

	var readSize int
	var err error
	var count int

	buf := make([]byte, 1024)

	for {
		readSize, err = r.Read(buf)
		if err != nil {
			break
		}

		var buffPosition int
		for {
			i := bytes.IndexByte(buf[buffPosition:], '\n')
			if i == -1 || readSize == buffPosition {
				break
			}
			buffPosition += i + 1
			count++
		}
	}
	if readSize > 0 && count == 0 || count > 0 {
		count++
	}
	if err == io.EOF {
		return count - 1, nil
	}

	return count - 1, err
}

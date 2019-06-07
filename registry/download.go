package registry

import (
	"archive/zip"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type WriteCounter struct {
	Current uint64
	Total   uint64
	B       *mpb.Bar
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	//wc.Current += uint64(n)
	//wc.PrintProgress()
	wc.B.IncrBy(int(n))
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s of %s", humanize.Bytes(wc.Current), humanize.Bytes(wc.Total))
}

func DownloadFile(wg *sync.WaitGroup, dest string, url string, prg *mpb.Progress) { //error {
	defer wg.Done()

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(dest + ".tmp")
	if err != nil {
		//return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return //err
	}
	defer resp.Body.Close()

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	checkErr(err)

	task := filepath.Base(dest)
	job := " downloading "
	bar := prg.AddBar(int64(size),
		mpb.PrependDecorators(
			decor.Name(task, decor.WC{W: len(task) + 1, C: decor.DSyncWidthR}),
			decor.Name(job, decor.WC{W: len(job) + 1, C: decor.DSyncWidth}),
			decor.CountersKibiByte("% 6.1f / % 6.1f"),
			//decor.OnComplete(decor.CountersKibiByte("% 6.1f / % 6.1f"), "Done!"),
		),
		mpb.AppendDecorators(decor.Percentage(decor.WC{W: 5})),
	)

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{Total: uint64(size), B: bar}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		//return err
	}

	err = os.Rename(dest+".tmp", dest)
	checkErr(err)

	var filenames []string

	r, err := zip.OpenReader(dest)
	checkErr(err)
	//if err != nil {
	//	return filenames, err
	//}
	defer r.Close()

	for _, f := range r.File {

		rc, err := f.Open()
		checkErr(err)
		//if err != nil {
		//	return filenames, err
		//}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(filepath.Dir(dest), f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(filepath.Dir(dest))+string(os.PathSeparator)) {
			//return filenames, fmt.Errorf("%s: illegal file path", fpath)
			return //fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				checkErr(err)
				//return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				checkErr(err)
				//return filenames, err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			checkErr(err)
			//if err != nil {
			//	return filenames, err
			//}

		}
	}
	//return filenames, nil

	// Do Import here

	return //nil
}

func CreateReleasableUrl(year string) string {
	if year == "" {
		return US_RELEASABLE_AIRCRAFT_CURRENT
	} else {
		return fmt.Sprintf(US_RELEASABLE_AIRCRAFT_YEARLY_TEMPLATE, year)
	}
}

func checkErr(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}

// @license
// Copyright (C) 2018  Dinko Korunic
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"fmt"
	"github.com/karrick/godirwalk"
	"github.com/pborman/getopt/v2"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

const testDirName = "findlargedir"
const defaultAlertThreshold = 50000
const defaultTestFileCount = 20000
const defaultProgressTicker = time.Minute * 5

var alertThreshold, testFileCount *int64
var helpFlag, accurateFlag, progressFlag, isilonFlag, cloexecFlag *bool

func init() {
	alertThreshold = getopt.Int64Long("threshold", 't', defaultAlertThreshold,
		fmt.Sprintf("set file count threshold for alerting (default %v)", defaultAlertThreshold))
	testFileCount = getopt.Int64Long("testcount", 'c', defaultTestFileCount,
		fmt.Sprintf("set initial file count for inode size testing phase (default %v)", defaultTestFileCount))
	helpFlag = getopt.BoolLong("help", 'h', "display help")
	accurateFlag = getopt.BoolLong("accurate", 'a', "full accuracy when checking large directories")
	progressFlag = getopt.BoolLong("progress", 'p', "display progress status every 5 minutes")
	isilonFlag = getopt.BoolLong("isilon", '7', "enable support for EMC Isilon OneFS 7.x")
	cloexecFlag = getopt.BoolLong("cloexec", 'x', "disable open O_CLOEXEC for really ancient Unix systems")
}

func main() {
	getopt.Parse()
	args := getopt.Args()

	if *helpFlag || len(args) < 1 {
		getopt.PrintUsage(os.Stderr)
		os.Exit(0)
	}

	log.Printf("Note: program will attempt to identify directories larger than %v entries. Make sure you have r/w privileges.",
		*alertThreshold)

	// If Unix system doesn't support open O_CLOEXEC, try monkey patching syscall.Open
	// This will work only on FreeBSD and derivatives
	if *cloexecFlag {
		patchSyscallOpen()
	}

	// If using EMC Isilon 7.x compatibility, try monkey patching syscall.Stat and syscall.Lstat
	if *isilonFlag {
		patchSyscallStat()
		patchSyscallLstat()
	}

	for i := range args {
		processDirectory(args[i])
	}
}

// processDirectory will process individual root filesystem/folder path and identify blackhole directory offenders
func processDirectory(processPath string) {
	// Establish file to direcory inode ratio
	ratio := getInodeRatio(processPath)
	if ratio <= 0 {
		log.Printf("Unable to calculate inode to file count ratio on %q. Skipping.", processPath)
		return
	}

	var wg sync.WaitGroup

	var ticker *time.Ticker
	var lastPathname *string
	doneTickerChan := make(chan struct{}, 1)
	defer close(doneTickerChan)

	// Default 5-minute progress update if progressFlag is true
	if *progressFlag {
		ticker = time.NewTicker(defaultProgressTicker)

		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ticker.C:
					printPath(lastPathname)
				case <-doneTickerChan:
					ticker.Stop()
					return
				}
			}
		}()
	}

	signalChan := make(chan os.Signal, 1)
	doneSignalChan := make(chan struct{}, 1)
	defer close(doneSignalChan)

	// Default SIGUSR1 and SIGUSR2 handler which will display progress update
	registerStatusSignal(signalChan)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-signalChan:
				printPath(lastPathname)
			case <-doneSignalChan:
				return
			}
		}
	}()

	var offenderTotal, countFromStat int64

	// Fast concurrent directory walker: won't follow symlinks and won't sort entries
	godirwalk.Walk(processPath, &godirwalk.Options{
		Unsorted:            true,
		FollowSymbolicLinks: false,
		// Default callback will process only directory entries
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if de.IsDir() {
				// Process only if entry is directory
				lastPathname = &osPathname
				fi, err := os.Stat(osPathname)
				if err != nil {
					return err
				}

				countFromStat = int64(float64(fi.Size()) / ratio)
				if countFromStat >= int64(*alertThreshold) {
					log.Printf("Directory %q is possibly a large directory with %v entries.", osPathname,
						humanPrint(countFromStat))
					offenderTotal++

					// If necessary deep-dive the directory and get accurate file count
					if *accurateFlag {
						log.Printf("Calculating %q directory exact entry count. Please wait...",
							osPathname)

						deChildren, err := godirwalk.ReadDirents(osPathname, nil)
						if err != nil {
							log.Print(err)
							return err
						}

						log.Printf("Done. Directory %q has exactly %v entries.", osPathname,
							len(deChildren))
					}
					return fmt.Errorf("directory %q is too large to process", osPathname)
				}
			}
			return nil
		},
		// Default error callback will just skip over when encountering errors
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			return godirwalk.SkipNode
		},
	})

	// Close channels and cleanup routines
	if *progressFlag {
		doneTickerChan <- struct{}{}
	}
	doneSignalChan <- struct{}{}
	wg.Wait()

	log.Printf("Found %v large directories in %q.", offenderTotal, processPath)
}

// humanPrint will display base10 approximate file count
func humanPrint(input int64) string {
	exp := math.Round(math.Log(float64(input)) / math.Log(float64(10)))

	switch {
	case exp >= 10:
		return "~10b"
	case exp >= 9:
		return "~1b"
	case exp >= 8:
		return "~100m"
	case exp >= 7:
		return "~10m"
	case exp >= 6:
		return "~1m"
	case exp >= 5:
		return "~100k"
	case exp >= 4:
		return "~10k"
	case exp >= 3:
		return "~1k"
	}

	return "<1k"
}

// printPath will display path processing progress
func printPath(processPath *string) {
	if processPath != nil && *processPath != "" {
		log.Printf("Last processed path was: %q.", *processPath)
	}
}

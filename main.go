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
	"log"
	"math"
	"os"

	"github.com/karrick/godirwalk"
)

const testFileCount = 10000
const testDirName = "findlargedir"
const fileThreshold = 50000

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: specify one or more directories to scan for large subdirectories. Make sure you have r/w permissions.")
	}

	log.Printf("Note: program will attempt to alert on directories larger than %v entries by default.", fileThreshold)

	dirs := os.Args[1:]

	for i := range dirs {
		var offenderTotal, countFromStat int64

		if ratio := getInodeRatio(dirs[i]); ratio > 0 {
			godirwalk.Walk(dirs[i], &godirwalk.Options{
				Unsorted:            true,
				FollowSymbolicLinks: false,
				Callback: func(osPathname string, de *godirwalk.Dirent) error {
					if de.IsDir() {
						fi, err := os.Stat(osPathname)
						if err != nil {
							return err
						}

						countFromStat = int64(fi.Size() / ratio)
						if countFromStat > int64(fileThreshold) {
							log.Printf("Directory %q is possibly a large directory with %v entries.", osPathname,
								humanPrint(countFromStat))
							offenderTotal++
							return fmt.Errorf("directory %q is too large to process", osPathname)
						}
					}
					return nil
				},
				ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
					return godirwalk.SkipNode
				},
			})
		}
		log.Printf("Found %v large directories in %q.", offenderTotal, dirs[i])
	}
}

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

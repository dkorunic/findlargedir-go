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
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/dkorunic/findlargedir/cerrgroup"
	"github.com/dkorunic/findlargedir/isilonstat"
)

const testContent = "Death is lighter than a feather, but Duty is heavier than a mountain."
const minRatio = 1
const maxRatio = 128

// getInodeRatio will do a rough estimation on how much a single file occupies in a directory inode
func getInodeRatio(checkDir string) (ratio float64) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Errors encountered, skipping directory scan on %q.", checkDir)
			ratio = 0
		}
	}()

	log.Printf("Determining inode to file count ratio on %q. Please wait, creating %v files...", checkDir,
		*testFileCount)

	// Create a temporary directory in each root filesystem path and remove on exit
	tempDir, err := ioutil.TempDir(checkDir, testDirName)
	if err != nil {
		log.Print(err)
		return
	}
	defer os.RemoveAll(tempDir)

	// Get empty directory inode size
	dirSizeEmpty, err := getDirSize(tempDir)
	if err != nil {
		log.Print(err)
		return
	}

	// Highly concurrent file creation routine with at most NumCPU() running routines
	cg := cerrgroup.New(runtime.NumCPU())
	content := []byte(testContent)
	for i := int64(0); i < *testFileCount; i++ {
		cg.Go(func() error {
			t, err := ioutil.TempFile(tempDir, "")
			if err != nil {
				log.Print(err)
				return err
			}

			if _, err := t.Write(content); err != nil {
				log.Print(err)
				return err
			}

			if err := t.Close(); err != nil {
				log.Print(err)
				return err
			}

			return nil
		})
	}

	// Wait for all routines to finish
	if err = cg.Wait(); err != nil {
		log.Print(err)
		return
	}

	// Get full directory inode size
	dirSizeFull, err := getDirSize(tempDir)
	if err != nil {
		log.Print(err)
		return
	}

	// Stat st_size value sanity check
	if dirSizeFull < (minRatio**testFileCount) || dirSizeFull > (maxRatio**testFileCount) {
		log.Printf("Directory stat st_size structure is most likely incorrect (%v bytes used). Skipping folder checks.",
			dirSizeFull)
		return
	}

	// Calculate final file inode usage ratio
	ratio = float64(dirSizeFull-dirSizeEmpty) / float64(*testFileCount)

	// Ratio sanity check
	if ratio < minRatio || ratio > maxRatio {
		log.Printf("Calculated ratio (%v) failed sanity checking. Skipping folder checks.", ratio)
		ratio = 0
		return
	}

	log.Printf("Done. Approximate directory inode size to file count ratio on %q is %v.", checkDir, ratio)
	return
}

// getDirSize returns inode size from Fileinfo structure
func getDirSize(name string) (int64, error) {
	if *isilonFlag {
		fi, err := isilonstat.Stat(name)
		return fi.Size(), err
	}

	fi, err := os.Stat(name)
	return fi.Size(), err
}

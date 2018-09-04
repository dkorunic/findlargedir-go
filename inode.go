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
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"log"
	"os"
)

const testContent = "Death is lighter than a feather, but Duty is heavier than a mountain."

var g errgroup.Group

func getInodeRatio(checkDir string) (ratio float64) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Errors encountered, skipping directory scan on %q.", checkDir)
			ratio = 0
		}
	}()

	log.Printf("Determining inode to file count ratio on %q. Please wait, creating %v files...", checkDir,
		*testFileCount)

	tempDir, err := ioutil.TempDir(checkDir, testDirName)
	if err != nil {
		log.Print(err)
		return
	}
	defer os.RemoveAll(tempDir)

	fiEmpty, err := os.Stat(tempDir)
	if err != nil {
		log.Print(err)
		return
	}

	content := []byte(testContent)
	for i := int64(0); i < *testFileCount; i++ {
		g.Go(func() error {
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

	if err = g.Wait(); err != nil {
		log.Print(err)
		return
	}

	fiFull, err := os.Stat(tempDir)
	if err != nil {
		log.Print(err)
		return
	}

	ratio = float64(fiFull.Size()-fiEmpty.Size()) / float64(*testFileCount)
	log.Printf("Done. Approximate directory inode size to file count ratio on %q is %v.", checkDir, ratio)
	return
}

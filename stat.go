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

// +build freebsd,amd64

package main

import (
	"github.com/dkorunic/findlargedir/isilonstat"
	"github.com/dkorunic/findlargedir/monkey"
	"log"
	"os"
)

// patchSyscallStat will attempt to monkey patch syscall.Stat with our Isilon version
func patchSyscallStat() {
	log.Print("Attempting to monkey patch syscall.Stat. We might horribly crash here...")
	monkey.Patch(os.Stat, isilonstat.Stat)
	log.Print("Patching syscall.Stat done.")
}

// patchSyscallLstat will attempt to monkey patch syscall.Lstat with our Isilon version
func patchSyscallLstat() {
	log.Print("Attempting to monkey patch syscall.Lstat. We might horribly crash here...")
	monkey.Patch(os.Lstat, isilonstat.Lstat)
	log.Print("Patching syscall.Lstat done.")
}

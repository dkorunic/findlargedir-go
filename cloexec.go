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
	"fmt"
	"github.com/dkorunic/findlargedir/monkey"
	"log"
	"syscall"
	"unsafe"
)

// patchSyscallOpen will attempt to monkey patch syscall.Open and avoid using O_CLOEXEC
func patchSyscallOpen() {
	log.Print("Attempting to monkey patch syscall.Open. We might horribly crash here...")
	monkey.Patch(syscall.Open, syscallOpenNoCloexec)
	log.Print("Patching syscall.Open done.")
}

func syscallOpenNoCloexec(path string, mode int, perm uint32) (fd int, err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(path)
	if err != nil {
		return
	}

	// Strip syscall.O_CLOEXEC
	if mode&syscall.O_CLOEXEC != 0 {
		mode &^= syscall.O_CLOEXEC
	}

	r0, _, e1 := syscall.Syscall(syscall.SYS_OPEN, uintptr(unsafe.Pointer(_p0)), uintptr(mode), uintptr(perm))
	fd = int(r0)
	if e1 != 0 {
		err = fmt.Errorf("syscall.SYS_OPEN: %s", e1)
	}
	return
}

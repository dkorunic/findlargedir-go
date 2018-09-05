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

// +build !windows,!arm64

// Package isilonstat provides Isilon OneFS 7.1-compatible stat(). This is a
// fairly naive and incomplete implementation. It might crash and malfunction in
// most horrible ways.

package isilonstat

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

type IsilonStat_t struct {
	Dev           uint32
	Ino           uint64 // FreeBSD7: uint32
	Mode          uint32 // FreeBSD7: uint16
	Nlink         uint16
	Uid           uint32
	Gid           uint32
	Rdev          uint32
	Atimespec     syscall.Timespec
	Mtimespec     syscall.Timespec
	Ctimespec     syscall.Timespec
	Size          int64
	Blocks        int64
	Blksize       int32
	Flags         uint32
	Gen           uint32
	Lspare        int32
	Birthtimespec syscall.Timespec
}

type fileStat struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	sys     IsilonStat_t
}

func (fs *fileStat) Size() int64        { return fs.size }
func (fs *fileStat) Mode() os.FileMode  { return fs.mode }
func (fs *fileStat) ModTime() time.Time { return fs.modTime }
func (fs *fileStat) Sys() interface{}   { return &fs.sys }
func (fs *fileStat) Name() string       { return fs.name }
func (fs *fileStat) IsDir() bool        { return fs.Mode().IsDir() }

func Stat(name string) (os.FileInfo, error) {
	var fs fileStat
	err := syscallIsilonStat(name, &fs.sys)
	if err != nil {
		return nil, &os.PathError{"isilonstat", name, err}
	}
	fillFileStatFromSys(&fs, name)
	return &fs, nil
}

func fillFileStatFromSys(fs *fileStat, name string) {
	fs.name = basename(name)
	fs.size = fs.sys.Size
	fs.modTime = timespecToTime(fs.sys.Mtimespec)
	fs.mode = os.FileMode(fs.sys.Mode & 0777)
	switch fs.sys.Mode & syscall.S_IFMT {
	case syscall.S_IFBLK:
		fs.mode |= os.ModeDevice
	case syscall.S_IFCHR:
		fs.mode |= os.ModeDevice | os.ModeCharDevice
	case syscall.S_IFDIR:
		fs.mode |= os.ModeDir
	case syscall.S_IFIFO:
		fs.mode |= os.ModeNamedPipe
	case syscall.S_IFLNK:
		fs.mode |= os.ModeSymlink
	case syscall.S_IFREG:
		// nothing to do
	case syscall.S_IFSOCK:
		fs.mode |= os.ModeSocket
	}
	if fs.sys.Mode&syscall.S_ISGID != 0 {
		fs.mode |= os.ModeSetgid
	}
	if fs.sys.Mode&syscall.S_ISUID != 0 {
		fs.mode |= os.ModeSetuid
	}
	if fs.sys.Mode&syscall.S_ISVTX != 0 {
		fs.mode |= os.ModeSticky
	}
}

func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

func basename(name string) string {
	i := len(name) - 1
	// Remove trailing slashes
	for ; i > 0 && name[i] == '/'; i-- {
		name = name[:i]
	}
	// Remove leading directory name
	for i--; i >= 0; i-- {
		if name[i] == '/' {
			name = name[i+1:]
			break
		}
	}

	return name
}

func syscallIsilonStat(name string, buf *IsilonStat_t) (err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(name)
	if err != nil {
		return
	}
	_, _, e := syscall.Syscall(syscall.SYS_STAT, uintptr(unsafe.Pointer(_p0)), uintptr(unsafe.Pointer(buf)), 0)
	if e != 0 {
		return fmt.Errorf("syscall.SYS_STAT: %s", e)
	}

	return nil
}

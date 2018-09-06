//+build !windows

package monkey

import (
	"golang.org/x/sys/unix"
)

// this function is super unsafe
// aww yeah
// It copies a slice to a raw memory location, disabling all memory protection before doing so.
func copyToLocation(location uintptr, data []byte) {
	f := rawMemoryAccess(location, len(data))

	page := rawMemoryAccess(pageStart(location), unix.Getpagesize())
	err := unix.Mprotect(page, unix.PROT_READ|unix.PROT_WRITE|unix.PROT_EXEC)
	if err != nil {
		panic(err)
	}
	copy(f, data[:])

	err = unix.Mprotect(page, unix.PROT_READ|unix.PROT_EXEC)
	if err != nil {
		panic(err)
	}
}

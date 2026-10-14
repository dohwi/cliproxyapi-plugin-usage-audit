//go:build !windows

package store

import "os"

// openAppend opens name for appending (see open_windows.go for why this is a function).
func openAppend(name string, perm os.FileMode, create bool) (*os.File, error) {
	flag := os.O_APPEND | os.O_WRONLY
	if create {
		flag |= os.O_CREATE
	}
	return os.OpenFile(name, flag, perm)
}

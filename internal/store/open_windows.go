//go:build windows

package store

import (
	"os"
	"syscall"
)

// openAppend opens name for appending with FILE_SHARE_DELETE. Go's own os.OpenFile does
// not request that share mode, and Windows refuses to rename or replace a file while any
// handle without it is open. Log rotation renames the live file, so the writer must be
// opened this way for rotation to work at all on Windows.
func openAppend(name string, perm os.FileMode, create bool) (*os.File, error) {
	p, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: name, Err: err}
	}
	disposition := uint32(syscall.OPEN_EXISTING)
	if create {
		disposition = syscall.OPEN_ALWAYS
	}
	h, err := syscall.CreateFile(p,
		syscall.FILE_APPEND_DATA,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil, disposition, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: name, Err: err}
	}
	return os.NewFile(uintptr(h), name), nil
}

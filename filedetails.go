package main

import (
	"github.com/rjkroege/edwood/internal/file"
	"os"
)

type filedetails struct {
	name string
	info os.FileInfo
	hash file.Hash // Used to check if the file has changed on disk since loaded.
}

// UpdateInfo updates File's info to d if file hash hasn't changed.
func (f *filedetails) UpdateInfo(filename string, d os.FileInfo) error {
	h, err := file.HashFor(filename)
	if err != nil {
		return warnError(nil, "failed to compute hash for %v: %v", filename, err)
	}
	if h.Eq(f.hash) {
		f.info = d
	}
	return nil
}

package request

import (
	"path/filepath"
	"strings"
)

func datafileDir(mcdir, dataFileID string) string {
	pieces := strings.Split(dataFileID, "-")
	return filepath.Join(mcdir, pieces[1][0:2], pieces[1][2:4])
}

func datafilePath(mcdir, dataFileID string) string {
	return filepath.Join(datafileDir(mcdir, dataFileID), dataFileID)
}

// DataFilePath returns the path in the materials commons repo for a file with the given id.
func DataFilePath(mcdir, dataFileID string) string {
	return datafilePath(mcdir, dataFileID)
}

// DataFileDir returns the directory path for a file
func DataFileDir(mcdir, dataFileID string) string {
	return datafileDir(mcdir, dataFileID)
}

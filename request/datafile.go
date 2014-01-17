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

func DataFilePath(mcdir, dataFileID string) string {
	return datafilePath(mcdir, dataFileID)
}

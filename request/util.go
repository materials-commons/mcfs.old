package request

import (
	"github.com/materials-commons/base/schema"
	"os"
)

func datafileSize(mcdir, dataFileID string) int64 {
	path := datafilePath(mcdir, dataFileID)
	finfo, err := os.Stat(path)
	switch {
	case err == nil:
		return finfo.Size()
	case os.IsNotExist(err):
		return 0
	default:
		return -1
	}
}

func datafileLocationID(dataFile *schema.File) string {
	if dataFile.UsesID != "" {
		return dataFile.UsesID
	}

	return dataFile.ID
}

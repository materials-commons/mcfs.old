package request

import (
	"github.com/materials-commons/mcfs/base/mc"
	"github.com/materials-commons/mcfs/base/schema"
	"os"
)

func datafileSize(mcdir, dataFileID string) int64 {
	path := mc.FilePathFrom(mcdir, dataFileID)
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

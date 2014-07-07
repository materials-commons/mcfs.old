package ws

import (
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"

	"github.com/materials-commons/mcfs/base/mc"
)

func (r uploadResource) startAssemblers() {
	for i := 0; i < maxAssemblers; i++ {
		go r.fileAssembler()
	}
}

func (r uploadResource) fileAssembler() {
	for request := range r.assembleRequest {
		r.assembleFile(request)
	}
}

type byChunk []os.FileInfo

func (c byChunk) Len() int      { return len(c) }
func (c byChunk) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c byChunk) Less(i, j int) bool {
	chunkIName, _ := strconv.Atoi(c[i].Name())
	chunkJName, _ := strconv.Atoi(c[j].Name())
	return chunkIName < chunkJName
}

func (r uploadResource) assembleFile(request finishRequest) {

	// reassemble file
	filePath := mc.FilePath(request.fileID)
	fdst, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer fdst.Close()

	finfos, err := ioutil.ReadDir(request.uploadPath)
	if err != nil {
		return
	}

	sort.Sort(byChunk(finfos))
	for _, finfo := range finfos {
		fsrc, err := os.Open(chunkPath(request.uploadPath, finfo.Name()))
		if err != nil {
			return
		}
		io.Copy(fdst, fsrc)
		fsrc.Close()
	}
	os.RemoveAll(request.uploadPath)
}

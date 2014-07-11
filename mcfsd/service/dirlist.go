package service

import (
	"github.com/materials-commons/mcfs/dir"
	"github.com/materials-commons/mcfs/schema"
	"path/filepath"
	"sort"
)

// dirList is the state structure for building a sorted
// list of entries from a denorm table lookup.
type dirList struct {
	files []dir.FileInfo
}

// fileList type is defined for Sort.
type fileList []dir.FileInfo

// Len length of array for sort.
func (f fileList) Len() int {
	return len(f)
}

// Swap swaps elements for Sort.
func (f fileList) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// Less determines whether one item is less than another for Sort.
// We compare the lexographic order of the path.
func (f fileList) Less(i, j int) bool {
	return f[i].Path < f[j].Path
}

// build creates the sorted list of dir.FileInfo files and directories from a list
// of DataDirDenorm items.
func (dlist *dirList) build(denormEntries []schema.DataDirDenorm, base string) []dir.FileInfo {
	for _, d := range denormEntries {
		dlist.addDirectory(d, base)
	}

	sort.Sort(fileList(dlist.files))
	return dlist.files
}

// addDirectory adds the directory and all its files to the list of entries.
func (dlist *dirList) addDirectory(d schema.DataDirDenorm, base string) {
	// Add the directory
	newDir := dir.FileInfo{
		ID:    d.ID,
		Path:  filepath.Join(base, d.Name),
		MTime: d.Birthtime,
		IsDir: true,
	}
	dlist.files = append(dlist.files, newDir)

	// Add directory's files
	for _, f := range d.DataFiles {
		newFile := dir.FileInfo{
			ID:       f.ID,
			Path:     filepath.Join(base, d.Name, f.Name),
			Size:     f.Size,
			Checksum: f.Checksum,
			MTime:    f.Birthtime,
		}
		dlist.files = append(dlist.files, newFile)
	}
}

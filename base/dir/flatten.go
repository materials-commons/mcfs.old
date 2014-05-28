package dir

// flattenState stores the state of the flatten progress.
type flattenState struct {
	all []FileInfo
}

// Flatten takes a Directory and flattens it into a list of file objects
// sorted by full path. It does this for the entire set of files, including
// files in sub directories.
func (d *Directory) Flatten() []FileInfo {
	state := &flattenState{
		all: []FileInfo{},
	}

	// Add top level directory
	state.all = append(state.all, d.FileInfo)

	// Now flatten the entries
	state.flatten(d)
	return state.all
}

// flatten does the actual work of flattening the directory files into
// a list and descending down through all the sub directories.
func (s *flattenState) flatten(d *Directory) {
	// The entries from the walk are already sorted lexographically from Walk.
	for _, file := range d.Files {
		s.all = append(s.all, file)
		if file.IsDir {
			s.flatten(d.SubDirectories[file.Path])
		}
	}
}

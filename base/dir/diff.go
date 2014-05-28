package dir

type diffState struct {
	origFiles []FileInfo
	newFiles  []FileInfo
	patches   []Patch
}

// Diff compares two versions of a directory tree over time. The first directory
// is the original or older version. The second directory is the new version.
// Diff creates a list of patches that will take the original version and
// transform it into the newer version.
func Diff(originalVersion *Directory, newerVersion *Directory) []Patch {
	origFiles := originalVersion.Flatten()
	newFiles := newerVersion.Flatten()
	return DiffFlat(origFiles, newFiles)
}

// DiffFlat compares two versions of a directory tree over time. It is like
// Diff except that it expects to receive flattened, sorted lists of the
// two directories contents.
func DiffFlat(origFiles, newFiles []FileInfo) []Patch {
	state := &diffState{
		origFiles: origFiles,
		newFiles:  newFiles,
		patches:   []Patch{},
	}

	state.computePatches()
	return state.patches
}

// computePatches creates a patch list of changes to apply to the original directory
// to make it look like the new version.
func (s *diffState) computePatches() {
	origLen := len(s.origFiles)
	newLen := len(s.newFiles)
	origIndex, newIndex := 0, 0

DIR_COMPARE_LOOP:
	for {
		switch {
		case origIndex >= origLen && newIndex >= newLen:
			break DIR_COMPARE_LOOP

		case origIndex >= origLen:
			// We are at the end of the list for origFiles any files in newFiles are
			// not in origFiles. This means that all these files were created since
			// the original version. Add a create patch.
			s.addPatch(s.newFiles[newIndex], PatchCreate)

		case newIndex >= newLen:
			// We are at the end of the list for newFiles. Any files in origFiles are not
			// in newFiles. This means that these files were deleted. Add a delete patch.
			s.addPatch(s.origFiles[origIndex], PatchDelete)

		case s.origFiles[origIndex].Path > s.newFiles[newIndex].Path:
			// There is a file in origFiles that is not in newFiles - add a delete patch.
			s.addPatch(s.origFiles[origIndex], PatchDelete)

			// Decrement origIndex because we are going to increment at the bottom
			// of the loop. Thus this decrement means we will be comparing this same origFiles
			// entry again against another entry in newFiles. We will keep doing this until
			// newFiles catches up or we run out of newFiles entries.
			origIndex--

		case s.origFiles[origIndex].Path < s.newFiles[newIndex].Path:
			// There is a file in newFiles that is not in origFiles - add a create patch.
			// from newFiles.
			s.addPatch(s.newFiles[newIndex], PatchCreate)

			// Decrement newIndex because we are going to increment at the bottom
			// of the loop. Thus this decrement means we will be comparing this same newFiles
			// entry again against another entry in origFiles. We will keep doing this until
			// origFiles catches up or we run out of origFiles entries.
			newIndex--

		default:
			// The file exists in both the new and the old versions. So we need to check
			// if it changed. This is a simple check of the ModTime. If the mod time is
			// different then there was some sort of change. This doesn't guarantee that
			// the contents have changed. If they haven't changed this will be picked up
			// by the upload process.
			origMTime := s.origFiles[origIndex].MTime
			newMTime := s.newFiles[newIndex].MTime
			if origMTime.Before(newMTime) {
				// File changed create an edit.
				s.addPatch(s.newFiles[newIndex], PatchEdit)
			}
		}
		origIndex++
		newIndex++
	}
}

// addPatch adds a patch entries to the state.
func (s *diffState) addPatch(f FileInfo, patchType PatchType) {
	patch := Patch{
		File: f,
		Type: patchType,
	}
	s.patches = append(s.patches, patch)
}

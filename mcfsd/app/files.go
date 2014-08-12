package app

import (
	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/dai"
)

// File represents a file in the system.
type File struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ProjectID   string `json:"project_id"`
	DirectoryID string `json:"directory_id"`
	Checksum    string `json:"checksum"`
	Owner       string `json:"owner"`
	Size        int64  `json:"size"`
	Projects    []OID  `json:"projects"`
}

// isValid sanity checks the entries in the file object. The filesService
// will check the existence of key entries such as the directory.
func (f File) isValid() error {
	switch {
	case f.Size < 1:
		return mcerr.Errorf(mcerr.ErrInvalid, "Invalid size (%d)", f.Size)
	case f.Checksum == "":
		return mcerr.Errorf(mcerr.ErrInvalid, "No checksum")
	case f.Owner == "":
		return mcerr.Errorf(mcerr.ErrInvalid, "No owner")
	case f.Name == "":
		return mcerr.Errorf(mcerr.ErrInvalid, "No name")
	case f.ProjectID == "":
		return mcerr.Errorf(mcerr.ErrInvalid, "No project")
	case f.DirectoryID == "":
		return mcerr.Errorf(mcerr.ErrInvalid, "No directory")
	default:
		return nil
	}
}

// FilesService represents the application operations on a file.
type FilesService interface {
	Create(file File) (*schema.File, error)
}

// filesService is the concrete representation of the FilesService interface.
type filesService struct {
	files    dai.Files
	dirs     dai.Dirs
	projects dai.Projects
	groups   dai.Groups
}

// NewFilesService returns a new FilesService.
func NewFilesService(files dai.Files, dirs dai.Dirs, projects dai.Projects, groups dai.Groups) *filesService {
	return &filesService{
		files:    files,
		dirs:     dirs,
		projects: projects,
		groups:   groups,
	}
}

// Create will create a new file in the repo or return an existing file. If a
// new file is created it returns the file entry and sets error to nil. If
// an existing file is returned, then error is set to mcerr.ErrExists. Any other
// error means that no entry was found or created.
func (s *filesService) Create(file File) (*schema.File, error) {

	if err := s.validate(file); err != nil {
		return nil, err
	}

	return s.createFile(file)
}

// validate will validate the entries for the file. It checks that values look reasonable
// and the project and directory exist.
func (s *filesService) validate(file File) error {
	if err := file.isValid(); err != nil {
		return err
	}

	proj, err := s.projects.ByID(file.ProjectID)
	if err != nil {
		return mcerr.Errorf(mcerr.ErrInvalid, "Bad project id %s", file.ProjectID)
	}

	if !s.groups.HasAccess(proj.Owner, file.Owner) {
		return mcerr.ErrNoAccess
	}

	ddir, err := s.dirs.ByID(file.DirectoryID)
	if err != nil {
		return mcerr.Errorf(mcerr.ErrInvalid, "Unknown directory id: %s", file.DirectoryID)
	}

	if ddir.Project != file.ProjectID {
		return mcerr.Errorf(mcerr.ErrInvalid, "Directory %s not in project %s", ddir.Name, file.ProjectID)
	}

	return nil
}

func (s *filesService) createFile(file File) (*schema.File, error) {
	// Check if the file exists, and if so how many versions. Do this within
	// the context of a particular directory, name and checksum. You can end
	// up with multiple matches on a checksum if a file keeps going back
	// and forth between two different versions of a file.
	files, err := s.files.ByPathChecksum(file.Name, file.DirectoryID, file.Checksum)
	switch {
	case err != nil:
		return nil, err
	case len(files) == 0:
		// This is the easy case. No matching files were found, so we just create a
		// new file. There may be an existing current file with a different checksum
		// so we need to handle that case as well.
		return s.createNewFile(file)

	case len(files) == 1:
		// Only one match. We have either a partial, or a fully uploaded file.
		// In either case return the existing entry and let the call know its
		// an existing file.
		return &files[0], mcerr.ErrExists

	default:
		// There are multiple matches. That means we could have old versions,
		// or a partial. Lets see if there is a partial. If there is then
		// this is easy, we just return the partial. If there isn't then the
		// checksum could match the current file. Return that if true, otherwise
		// we need to create a new file version.
		current := schema.Files.Find(files, func(f schema.File) bool { return f.Current })
		partial := schema.Files.Find(files, func(f schema.File) bool { return f.Size != f.Uploaded && !f.Current })
		switch {
		case partial != nil:
			// There is an existing partial no need to create a new file.
			return partial, mcerr.ErrExists

		case current != nil:
			// The checksum matches the current file.
			return current, mcerr.ErrExists

		default:
			// No partial, and no match on existing. Create a new file.
			return s.createNewFile(file)
		}
	}
}

// createNewFile will create the file object in the database. It inserts a new file entry
// but doesn't attach it up to dependent objects. We assume that the file still needs
// book keeping, such as being uploaded.
func (s *filesService) createNewFile(file File) (*schema.File, error) {
	var f *schema.File

	currentFile, err := s.files.ByPath(file.Name, file.DirectoryID)
	switch {
	case err == mcerr.ErrNotFound:
		// There is no current entry, create a new one.
		f = s.newFile(file)
	case err != nil:
		// Database error occurred.
		return nil, err
	default:
		// There is a current entry, so create the new one with the parent pointing
		// to the current entry.
		f = s.newFile(file)
		f.Parent = currentFile.ID
	}

	created, err := s.files.InsertEntry(f)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// newFile creates a new file object to insert into the database. It also handles the
// bookkeeping task of setting the usesid field if their is a file matching this
// checksum that has already been uploaded.
func (s *filesService) newFile(file File) *schema.File {
	f := schema.NewFile(file.Name, file.Owner)
	f.DataDirs = append(f.DataDirs, file.DirectoryID)
	f.Checksum = file.Checksum
	f.Size = file.Size
	f.Current = false

	dup, err := s.files.ByChecksum(file.Checksum)
	if err == nil && dup != nil {
		// Found a matching entry, set usesid to it
		f.UsesID = dup.ID
	}

	return &f
}

package materials

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/materials-commons/gohandy/file"
	"os"
	"path/filepath"
	"time"
)

var (
	// ErrBadProjectFileStatusString status string is unknown.
	ErrBadProjectFileStatusString = fmt.Errorf("unknown string value for ProjectFileStatus")

	// ErrBadProjectFileLocationString file location is unknown
	ErrBadProjectFileLocationString = fmt.Errorf("unknown string value for ProjectFileLocation")
)

// ProjectFileStatus is the state of a file in the project.
type ProjectFileStatus int

// UnsetOption unsets the option.
const UnsetOption = 0

const (
	// Synchronized the file has been synchronized with the server
	Synchronized ProjectFileStatus = iota

	// Unsynchronized the file hasn't been synched with the server
	Unsynchronized

	// New the file is new on the client
	New

	// Deleted the file has been deleted
	Deleted

	// UnknownFileStatus we can't determine the files status
	UnknownFileStatus
)

var pfs2Strings = map[ProjectFileStatus]string{
	Synchronized:      "Synchronized",
	Unsynchronized:    "Unsynchronized",
	New:               "New",
	Deleted:           "Deleted",
	UnknownFileStatus: "UnknownFileStatus",
}

var pfsString2Value = map[string]ProjectFileStatus{
	"Synchronized":      Synchronized,
	"Unsynchronized":    Unsynchronized,
	"New":               New,
	"Deleted":           Deleted,
	"UnknownFileStatus": UnknownFileStatus,
}

// String implements the string interface for ProjectFileStatus
func (pfs ProjectFileStatus) String() string {
	str, found := pfs2Strings[pfs]
	switch found {
	case true:
		return str
	default:
		return "Unknown"
	}
}

// String2ProjectFileStatus converts a string to a ProjectFileStatus
func String2ProjectFileStatus(pfs string) (ProjectFileStatus, error) {
	val, found := pfsString2Value[pfs]
	switch found {
	case true:
		return val, nil
	default:
		return -1, ErrBadProjectFileStatusString
	}
}

// ProjectFileLocation is the location of a file.
type ProjectFileLocation int

const (
	// LocalOnly the file exists only on the local client
	LocalOnly ProjectFileLocation = iota

	// RemoteOnly the file exists only on the server
	RemoteOnly

	// LocalAndRemote the file exists both on the local client and the server
	LocalAndRemote

	// LocalAndRemoteUnknown the exact location of the file cannot be determined
	LocalAndRemoteUnknown
)

var pfl2Strings = map[ProjectFileLocation]string{
	LocalOnly:             "LocalOnly",
	RemoteOnly:            "RemoteOnly",
	LocalAndRemote:        "LocalAndRemote",
	LocalAndRemoteUnknown: "LocalAndRemoteUnknown",
}

var pflString2Value = map[string]ProjectFileLocation{
	"LocalOnly":             LocalOnly,
	"RemoteOnly":            RemoteOnly,
	"LocalAndRemote":        LocalAndRemote,
	"LocalAndRemoteUnknown": LocalAndRemoteUnknown,
}

// String implements the Stringer interface for ProjectFileLocation.
func (pfl ProjectFileLocation) String() string {
	str, found := pfl2Strings[pfl]
	switch found {
	case true:
		return str
	default:
		return "Unknown"
	}
}

// String2ProjectFileLocation converts a string to ProjectFileLocation
func String2ProjectFileLocation(pfl string) (p ProjectFileLocation, err error) {
	val, found := pflString2Value[pfl]
	switch found {
	case true:
		return val, nil
	default:
		return -1, ErrBadProjectFileLocationString
	}
}

// ProjectFileInfo holds all the information on a project file.
type ProjectFileInfo struct {
	Path     string
	Size     int64
	Hash     string
	ModTime  time.Time
	ID       string
	Status   ProjectFileStatus
	Location ProjectFileLocation
}

// TrackingOptions describes a files status and location.
type TrackingOptions struct {
	FileStatus   ProjectFileStatus
	FileLocation ProjectFileLocation
}

// Walk walks a project and determines status.
func (project *Project) Walk(options *TrackingOptions) error {
	fileStatus := Unsynchronized
	fileLocation := LocalOnly

	if options != nil {
		if options.FileStatus != UnsetOption {
			fileStatus = options.FileStatus
		}

		if options.FileLocation != UnsetOption {
			fileLocation = options.FileLocation
		}
	}

	filepath.Walk(project.Path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			key := md5.New().Sum([]byte(path))
			checksum, _ := file.Hash(md5.New(), path)
			pinfo := &ProjectFileInfo{
				Path:     path,
				Size:     info.Size(),
				Hash:     fmt.Sprintf("%x", checksum),
				ModTime:  info.ModTime(),
				Status:   fileStatus,
				Location: fileLocation,
			}
			value, _ := json.Marshal(pinfo)
			project.Put(key, value, nil)
		}
		return nil
	})

	return nil
}

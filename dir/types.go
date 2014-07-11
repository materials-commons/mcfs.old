package dir

import (
	"os"
	"time"
)

// FileInfo describes a file or directory entry
type FileInfo struct {
	ID       string    // ID of file/directory
	Path     string    // Full path including name
	Size     int64     // Size valid only for file
	Checksum string    // MD5 Hash - valid only for files
	MTime    time.Time // Modification time
	IsDir    bool      // True if this entry represents a directory
}

// newFile creates a new File entry.
func newFileInfo(path string, info os.FileInfo) FileInfo {
	fi := FileInfo{
		Path:  path,
		MTime: info.ModTime(),
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		fi.Size = info.Size()
	}

	return fi
}

// Directory is a container for the files and sub directories in a single directory.
// Each sub directory will itself contain a list of files and directories.
type Directory struct {
	FileInfo                             // Information about the directory
	Files          []FileInfo            // List of files and directories in this directory
	SubDirectories map[string]*Directory // List of directories in this directory
}

// newDirectory creates a new Directory entry.
func newDirectory(path string, info os.FileInfo) *Directory {
	return &Directory{
		FileInfo: FileInfo{
			Path:  path,
			MTime: info.ModTime(),
			IsDir: true,
		},
		SubDirectories: make(map[string]*Directory),
	}
}

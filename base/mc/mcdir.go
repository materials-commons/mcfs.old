package mc

import (
	"path/filepath"
	"strings"

	"github.com/materials-commons/config"
)

// FileDir returns the full directory path for a given fileID,
// using MCDIR as the base.
func FileDir(fileID string) string {
	return FileDirFrom(config.GetString("MCDIR"), fileID)
}

// FileDirFrom returns the full directory path for a given fileID
// using dir as the base.
func FileDirFrom(dir, fileID string) string {
	pieces := strings.Split(fileID, "-")
	return filepath.Join(dir, pieces[1][0:2], pieces[1][2:4])
}

// FilePath returns the full path for a given fileID using
// MCDIR as the base.
func FilePath(fileID string) string {
	return FilePathFrom(config.GetString("MCDIR"), fileID)
}

// FilePathFrom returns the full path for a given fileID
// using dir as the base.
func FilePathFrom(dir, fileID string) string {
	return filepath.Join(FileDirFrom(dir, fileID), fileID)
}

// Dir returns the base directory for MCDIR
func Dir() string {
	return config.GetString("MCDIR")
}

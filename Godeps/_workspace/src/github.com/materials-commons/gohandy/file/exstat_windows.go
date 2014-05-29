package file

import (
	"os"
	"path/filepath"
	"time"
)

// winExFileInfo stores windows specific file information.
// At the moment all the information we need is available
// through the Sys() interface.
type winExFileInfo struct {
	os.FileInfo
	fid  FID
	path string
}

// CTime returns the CreationTime from Win32FileAttributeData.
func (fi *winExFileInfo) CTime() time.Time {
	return time.Unix(0, fi.Sys().(syscall.Win32FileAttributeData).CreationTime)
}

// ATime returns the LastAccessTime from Win32FileAttributeData.
func (fi *winExFileInfo) ATime() time.Time {
	return time.Unix(0, fi.Sys().(syscall.Win32FileAttributeData).LastAccessTime)
}

// FID returns the windows version of a file id. The FID for Windows
// is the VolumeSerialNumber (IDHigh) and the FileIndexHigh/Low (IDLow)
func (fi *winExFileInfo) FID() FID {
	return fid
}

// Path returns the full path for the file.
func (fi *winExFileInfo) Path() string {
	return path
}

// newExFileInfo creates a new winExFileInfo from a os.FileInfo.
func newExFileInfo(fi os.FileInfo, path string) *winExFileInfo {
	fid, err := createFID(path)
	if err != nil {
		// do something
	}

	absolute, _ := filepath.Abs(path)
	return &winExFileInfo{
		FileInfo: fi,
		fid:      fid,
		path:     filepath.Clean(absolute),
	}
}

// createFID creates the file by making a windows specific system call
// to retrieve the VolumeSerialNumber and FileIndexHigh/Low. Unfortunately
// these values are not exposed through the Sys() in FileInfo. The code
// for making these calls is a slightly modified version of the code in
// the go os package types_windows.go file.
func createFID(path string) (FID, error) {
	fid := FID{}
	pathp, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return fid, err
	}
	h, err := syscall.CreateFile(pathp, 0, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return fid, err
	}
	defer syscall.CloseHandle(h)
	var handleInfo syscall.ByHandleFileInformation
	err = syscall.GetFileInformationByHandle(syscall.Handle(h), &handleInfo)
	if err != nil {
		return fid, err
	}
	fid.IDHigh = uint64(handleInfo.VolumeSerialNumber)
	fid.IDLow = int64(handleInfo.FileIndexHigh)<<32 + int64(handleInfo.FileIndexLow)
	return fid, nil
}
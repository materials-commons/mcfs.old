package dir

import (
	"fmt"
	"os"
	"path/filepath"
)

// walkerContext keeps contextual information around as a directory tree is walked.
// This allows the directory walk function to add items to the correct directory
// object.
type walkerContext struct {
	baseDir string                // Top level directory the walk was started at
	current *Directory            // Current directory being operated on
	all     map[string]*Directory // Every known directory
}

// Load walks a given directory path creating a Directory object with all files and
// sub directories filled out.
func Load(path string) (*Directory, error) {
	// Check if valid path
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	context, dir := createStartingState(path, fi)
	if err := filepath.Walk(path, context.directoryWalker); err != nil {
		return dir, err
	}
	return dir, nil
}

// createStartingState sets up the current context and top level directory. It sets
// context.current to this top level directory.
func createStartingState(path string, finfo os.FileInfo) (*walkerContext, *Directory) {
	dir := newDirectory(path, finfo)
	context := &walkerContext{
		current: dir,
		baseDir: path,
		all:     make(map[string]*Directory),
	}
	context.all[path] = dir
	return context, dir
}

// directoryWalker is the function called by filepath.Walk as it walks a directory tree.
// It builds up the list of entries, both files and directories, in each directory that
// is visited populating a top level Directory.
func (c *walkerContext) directoryWalker(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	c.setCurrentDir(path)

	if info.IsDir() {
		// Don't add the base directory as a sub directory. Otherwise our base
		// directory has itself as a sub directory in it's list of sub directories.
		if c.baseDir != path {
			c.addSubdir(path, info)
		}
	} else {
		c.addFile(path, info)
	}

	return nil
}

// setCurrentDir sets c.current to the directory being operated on. It checks
// the directory path for the current element and determines if current needs
// to be set to a different directory.
func (c *walkerContext) setCurrentDir(path string) {
	dirpath := filepath.Dir(path)
	if path != c.current.Path && dirpath != c.current.Path {
		// We have descended into a new directory, so set current to the
		// new entry.
		dir, ok := c.all[dirpath]
		if !ok {
			panic(fmt.Sprintf("Fatal: Could not find directory: %s", dirpath))
		}
		c.current = dir
	}
}

// addSubdir adds a new sub directory to the current directory.
func (c *walkerContext) addSubdir(path string, info os.FileInfo) {
	dir := newDirectory(path, info)
	c.current.SubDirectories[path] = dir
	c.current.Files = append(c.current.Files, dir.FileInfo)
	c.all[path] = dir
}

// addFile adds a new file to the current directory
func (c *walkerContext) addFile(path string, info os.FileInfo) {
	f := newFileInfo(path, info)
	c.current.Files = append(c.current.Files, f)
}

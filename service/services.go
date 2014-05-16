package service

//
var File Files

//
var Dir Dirs

//
var Project Projects

func init() {
	File = NewFiles(RethinkDB)
	Dir = NewDirs(RethinkDB)
	Project = NewProjects(RethinkDB)
}

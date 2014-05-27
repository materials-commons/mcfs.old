package tlog

import (
	"os"
)

type tlogServer struct {
	isRunning bool
	request chan interface{}
	response chan bool
	tlogFile *os.File
}

var server = &tlogServer{}

func Server() *tlogServer {
	return server
}

func (s *tlogServer) Append(what interface{}) {
	if !s.isRunning {
		panic("Not running")
	}
}

func (s *tlogServer) Init() {
	var err error
	s.request = make(chan interface{})
	s.response = make(chan bool)
	if s.tlogFile, err = os.OpenFile("", os.O_RDWR|os.O_APPEND, 0660); err != nil {
		// Panic here because we couldn't open the log
	}
}

func (s *tlogServer) Run(stopChan <-chan struct{}) {
	s.isRunning = true
	for {
		select {
		case request := <- s.request:
			s.writeEntry(request)
		case <-stopChan:
		}
	}
}

func (s *tlogServer) writeEntry(item interface{}) {
}





















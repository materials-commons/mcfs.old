package servers

import (
	"fmt"
	"testing"
	"time"
)

var _ = fmt.Println

type testServer struct {
	in  chan string
	out chan string
}

func (ts *testServer) Run(stopChan <-chan struct{}) {
	for {
		select {
		case request := <-ts.in:
			ts.out <- request
		case <-stopChan:
			close(ts.in)
			close(ts.out)
			return
		}
	}
}

func (ts *testServer) Init() {
	ts.in = make(chan string)
	ts.out = make(chan string)
}

var tserverInstance = &testServer{}

var tserver = Server{
	Instance: tserverInstance,
}

func TestStartStop(t *testing.T) {
	status := tserver.Status()
	if status != Stopped {
		t.Fatalf("Server not started, status should should be Stopped, got %s", status)
	}

	tserver.Start()
	status = tserver.Status()
	if status != Running {
		t.Fatalf("Server started, but shows status %s", status)
	}

	tserver.Stop()
	time.Sleep(1000 * 1)
	status = tserver.Status()
	if status != Stopped {
		t.Fatalf("Server stopped, but shows status %s", status)
	}
}

func TestCommunication(t *testing.T) {
	// Make sure that it starts and we can communicate with it.
	tserver.Start()
	tserverInstance.in <- "Hello"
	response := <-tserverInstance.out
	if response != "Hello" {
		t.Fatalf("Server running but got bad response: %s", response)
	}
}

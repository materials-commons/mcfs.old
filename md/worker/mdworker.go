//
//  Majordomo Protocol worker example.
//  Uses the mdwrk API to hide all MDP aspects
//

package worker

import (
	"github.com/materials-commons/mcfs/md/mdapi"

	"fmt"
	"log"
	"os"
)

func main() {
	var verbose = false
	var service = "echo"
	if len(os.Args) == 2 {
		service = os.Args[1]
	}
	fmt.Println("Starting worker for service:", service)
	session, _ := mdapi.NewMdwrk("tcp://localhost:5555", service, verbose)

	fmt.Println("work pid:", os.Getpid())
	var err error
	var request, reply []string
	for {
		request, err = session.Recv(reply)
		fmt.Println("  received for worker:", os.Getpid(), request)
		if err != nil {
			break //  Worker was interrupted
		}
		reply = request //  Echo is complex... :-)
	}
	log.Println(err)
}

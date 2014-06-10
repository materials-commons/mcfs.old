//
//  Majordomo Protocol client example - asynchronous.
//  Uses the mdcli API to hide all MDP aspects
//

package client2

import (
	"github.com/materials-commons/mcfs/md/mdapi"

	"fmt"
	"log"
	"os"
)

func main() {
	var verbose bool
	if len(os.Args) > 1 && os.Args[1] == "-v" {
		verbose = true
	}
	session, _ := mdapi.NewMdcli2("tcp://localhost:5555", verbose)

	var count int
	pid := os.Getpid()
	fmt.Println("Client pid:", pid)
	msg := fmt.Sprintf("Hello from %d", pid)
	for count = 0; count < 10; count++ {
		err := session.Send("echo", msg)
		if err != nil {
			log.Println("Send:", err)
			break
		}
	}
	for count = 0; count < 10; count++ {
		_, err := session.Recv()
		if err != nil {
			log.Println("Recv:", err)
			break
		}
	}
	fmt.Printf("%d replies received\n", count)
}

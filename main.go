package main

import (
	"log"

	"github.com/pebbe/zmq4"
)

func main() {
	context, err := zmq4.NewContext()
	if err != nil {
		log.Fatalln(err)
	}

	socket, err := context.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}

	if err := socket.Connect("tcp://127.0.0.1:5555"); err != nil {
		log.Fatalln(err)
	}

	for {
		reply, err := socket.Recv(0)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("Received reply [%s]\n", reply)
	}
}

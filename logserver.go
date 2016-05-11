package main

import (
	"LogServer/server"
	"log"
)

func main() {

	addr := "/tmp/logserver.sock"
	//addr := ":8080"
	log.Printf("Listen: trying to listen on %s\n", addr)

	err := server.Listen(addr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}

}

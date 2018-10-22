package main

import (
	"github.com/lugu/qiloop/bus/net"
	"github.com/lugu/qiloop/bus/session"
	"log"
)

func main() {
	listener, err := net.Listen("tcp://0.0.0.0:9559")
	if err != nil {
		log.Fatalf("%s", err)
	}
	directory := session.NewDirectoryService()
	router := session.NewRouter()
	router.Add(directory)
	server := session.NewServer(listener, router)
	server.Run()
}

package main

import (
	"github.com/lugu/qiloop/bus/net"
	"github.com/lugu/qiloop/bus/server"
	"github.com/lugu/qiloop/bus/server/directory"
	"log"
)

func main() {
	listener, err := net.Listen("tcp://0.0.0.0:9559")
	if err != nil {
		log.Fatalf("%s", err)
	}

	obj := directory.ServiceDirectoryObject(directory.NewServiceDirectory())
	srv := server.NewServer2(listener, server.NewRouter())
	srv.NewService("ServiceDirectory", obj)
	srv.Run()
}

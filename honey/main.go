package main

import (
	"github.com/lugu/qiloop/bus/directory"
	"log"
)

func main() {
	addr := "tcp://0.0.0.0:9559"
	dir, err := directory.NewServer(addr, nil)
	if err != nil {
		log.Fatalf("%s", err)
	}
	done := dir.WaitTerminate()
	<-done
}

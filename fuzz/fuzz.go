package fuzz

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/lugu/qiloop/bus"
	dir "github.com/lugu/qiloop/bus/directory"
	"github.com/lugu/qiloop/bus/net"
	"github.com/lugu/qiloop/bus/util"
)

var serverURL string = "tcps://localhost:9503"

func init() {
	serverURL = util.NewUnixAddr()
	server, err := dir.NewServer(serverURL, bus.Dictionary(
		map[string]string{
			"nao": "nao",
		},
	))

	go func() {
		err = <-server.WaitTerminate()
		if err != nil {
			log.Fatalf("Server: failed with error %s", err)
		}
	}()
}

func FuzzSerializer(data []byte) int {
	buf := bytes.NewBuffer(data)
	cm, err := bus.ReadCapabilityMap(buf)
	if err != nil {
		return 0
	}

	var out bytes.Buffer
	err = bus.WriteCapabilityMap(cm, &out)
	if err != nil {
		panic(err)
	}
	return 1
}

func Fuzz(data []byte) int {
	const serviceID = 0
	const objectID = 0
	const actionID = 8

	const timeout = 5 * time.Second

	endpoint, err := net.DialEndPoint(serverURL)
	if err != nil {
		log.Fatalf("failed to contact %s: %s", serverURL, err)
	}
	channel := bus.NewContext(endpoint)

	ch := make(chan bool, 1)
	defer close(ch)

	var err0 error
	go func() {
		ctx := context.Background()
		clt := bus.NewClient(channel)
		_, err0 = clt.Call(ctx.Done(), serviceID, objectID, actionID, data)
		ch <- true
	}()
	timer := time.NewTimer(timeout)

	select {
	case <-ch:
		timer.Stop()
		endpoint.Close()
	case <-timer.C:
		endpoint.Close()
		panic("gateway timeout1")
	}

	ch = make(chan bool, 1)

	go func() {
		// check if everything is still OK.
		endpoint, err = net.DialEndPoint(serverURL)
		if err != nil {
			panic("gateway has crashed")
		}
		err = bus.AuthenticateUser(endpoint, "nao", "nao")
		if err != nil {
			panic("gateway is broken")
		}
		ch <- true
	}()

	timer = time.NewTimer(timeout)
	select {
	case <-ch:
		timer.Stop()
		endpoint.Close()
	case <-timer.C:
		endpoint.Close()
		panic("gateway timeout2")
	}

	if err0 == nil {
		return 1
	}
	return 0
}

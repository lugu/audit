package fuzz

import (
	"bytes"
	"github.com/lugu/qiloop/bus"
	"github.com/lugu/qiloop/bus/net"
	"github.com/lugu/qiloop/type/value"
	"log"
)

var ServerURL string = "tcps://127.0.0.1:9503"

const serviceID = 0
const objectID = 0
const actionID = 8

func Fuzz(data []byte) int {

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

func Fuzz2(data []byte) int {
	endpoint, err := net.DialEndPoint(ServerURL)
	if err != nil {
		log.Fatalf("failed to contact %s: %s", ServerURL, err)
	}

	clt := bus.NewClient(endpoint)
	data, err0 := clt.Call(serviceID, objectID, actionID, data)

	// check response
	buf := bytes.NewBuffer(data)
	capability, err := bus.ReadCapabilityMap(buf)
	if err == nil {
		statusValue, ok := capability[bus.KeyState]
		if ok {
			status, ok := statusValue.(value.IntValue)
			if ok {
				switch uint32(status) {
				case bus.StateDone:
					panic("password found")
				case bus.StateContinue:
					panic("token renewal")
				}
			}
		}
	}

	// check if everything is still OK.
	endpoint2, err := net.DialEndPoint(ServerURL)
	if err != nil {
		panic("gateway has crashed")
	}

	err = bus.Authenticate(endpoint2)
	endpoint2.Close()
	if err != nil {
		panic("gateway is broken")
	}

	if err0 == nil {
		return 1
	}
	return 0
}

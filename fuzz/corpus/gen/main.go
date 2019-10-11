package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	gofuzz "github.com/google/gofuzz"
	"github.com/lugu/audit/fuzz"
	"github.com/lugu/qiloop/bus"
	"github.com/lugu/qiloop/bus/net"
	"github.com/lugu/qiloop/type/basic"
	"github.com/lugu/qiloop/type/object"
	"github.com/lugu/qiloop/type/value"
)

var ServerURL string = "tcps://localhost:9503"

const serviceID = 0
const objectID = 0
const actionID = 8

type objectReferenceValue struct {
	ref object.ObjectReference
}

func (o objectReferenceValue) Signature() string {
	return "o"
}

func (o objectReferenceValue) Write(w io.Writer) error {
	if err := basic.WriteString(o.Signature(), w); err != nil {
		return err
	}
	return object.WriteObjectReference(o.ref, w)
}

func (o objectReferenceValue) String() string {
	return fmt.Sprintf("%s, %s", o.Signature(), o.ref)
}

func makeValue(i *value.Value, c gofuzz.Continue) {
	switch c.Intn(12) {
	case 0:
		var b bool
		c.Fuzz(&b)
		*i = value.Bool(b)
	case 1:
		var u uint8
		c.Fuzz(&u)
		*i = value.Uint8(u)
	case 2:
		var u uint16
		c.Fuzz(&u)
		*i = value.Uint16(u)
	case 3:
		var u uint32
		c.Fuzz(&u)
		*i = value.Uint(u)
	case 4:
		var u int8
		c.Fuzz(&u)
		*i = value.Int8(u)
	case 5:
		var u int16
		c.Fuzz(&u)
		*i = value.Int16(u)
	case 6:
		var u int32
		c.Fuzz(&u)
		*i = value.Int(u)
	case 7:
		var f float32
		c.Fuzz(&f)
		*i = value.Float(f)
	case 8:
		var s string
		c.Fuzz(&s)
		*i = value.String(s)
	case 9:
		var list value.ListValue
		size := c.Intn(20)
		list = make([]value.Value, size)
		for i := 0; i < size; i++ {
			makeValue(&list[i], c)
		}
		*i = list
	case 10:
		*i = value.Void()
	case 11:
		var ref objectReferenceValue
		c.Fuzz(&ref)
		*i = ref
	}
}

func makeCap(f *gofuzz.Fuzzer) bus.CapabilityMap {
	var permission bus.CapabilityMap
	f.Fuzz(&permission)
	var user string
	f.Fuzz(&user)
	permission["user"] = value.String(user)
	var token string
	f.Fuzz(&token)
	permission["token"] = value.String(token)
	return permission
}

func Fuzz3() {
	fuzzer := gofuzz.New().NilChance(0).Funcs(makeValue).NumElements(1, 100)
	for i := 0; i < 10; i++ {
		perm := makeCap(fuzzer)
		fmt.Printf("%#v\n", perm)

		var buf bytes.Buffer
		err := bus.WriteCapabilityMap(perm, &buf)
		if err != nil {
			panic(err)
		}
		cm, err := bus.ReadCapabilityMap(&buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		fmt.Printf("%#v\n", cm)
	}
}

func authenticateCall(e net.EndPoint, p bus.CapabilityMap) error {

	cache := bus.NewCache(e)
	cache.AddService("ServiceZero", 0, object.MetaService0)
	proxies := bus.Services(cache)
	service0, err := proxies.ServiceServer()
	if err != nil {
		return err
	}

	resp, err := service0.Authenticate(p)
	if err != nil {
		return fmt.Errorf("authentication failed: %s", err)
	}

	statusValue, ok := resp[bus.KeyState]
	if !ok {
		return fmt.Errorf("missing authentication state")
	}
	status, ok := statusValue.(value.UintValue)
	if !ok {
		status2, ok := statusValue.(value.IntValue)
		if !ok {
			return fmt.Errorf("authentication status error (%#v)",
				statusValue)
		}
		status = value.UintValue(uint32(status2.Value()))
	}
	switch uint32(status) {
	case bus.StateDone:
		return nil
	case bus.StateContinue:
		return nil
	case bus.StateError:
		return fmt.Errorf("Authentication failed")
	default:
		return fmt.Errorf("invalid state type: %d", status)
	}
}

func Fuzz4() {

	fuzzer := gofuzz.New().NilChance(0).Funcs(makeValue).NumElements(1, 100)
	endpoint, err := net.DialEndPoint(ServerURL)
	if err != nil {
		log.Fatalf("failed to contact %s: %s", ServerURL, err)
	}
	defer endpoint.Close()
	perm := makeCap(fuzzer)
	fmt.Printf("%#v\n", perm)

	file, _ := ioutil.TempFile(".", "cap-*.bin")
	err = bus.WriteCapabilityMap(perm, file)
	if err != nil {
		panic(err)
	}
	file.Close()

	err = authenticateCall(endpoint, perm)
	if err == nil {
		log.Fatalf("%#v\n", perm)
	}
}

func main() {
	flag.StringVar(&ServerURL, "qi-url", ServerURL,
		"Service directory URL")
	flag.Parse()

	fuzz.WriteCorpus()

	for {
		Fuzz4()
	}
}

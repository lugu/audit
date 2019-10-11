//go:generate go run .

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"

	gofuzz "github.com/google/gofuzz"
	"github.com/lugu/audit/fuzz"
	"github.com/lugu/qiloop/bus"
	dir "github.com/lugu/qiloop/bus/directory"
	"github.com/lugu/qiloop/bus/net"
	"github.com/lugu/qiloop/bus/session/token"
	"github.com/lugu/qiloop/bus/util"
	"github.com/lugu/qiloop/type/basic"
	"github.com/lugu/qiloop/type/object"
	"github.com/lugu/qiloop/type/value"
)

var serverURL string = "tcps://localhost:9503"

func init() {
	serverURL = util.NewUnixAddr()
	user, token := token.GetUserToken()
	server, err := dir.NewServer(serverURL, bus.Dictionary(
		map[string]string{
			user: token,
		},
	))

	go func() {
		err = <-server.WaitTerminate()
		if err != nil {
			log.Fatalf("Server: failed with error %s", err)
		}
	}()
}

type objectReferenceValue struct {
	ref object.ObjectReference
}

func (o objectReferenceValue) Signature() string {
	return "(({I(Issss[(ss)<MetaMethodParameter,name,description>]s)<MetaMethod,uid,returnSignature,name,parametersSignature,description,parameters,returnDescription>}{I(Iss)<MetaSignal,uid,name,signature>}{I(Iss)<MetaProperty,uid,name,signature>}s)<MetaObject,methods,signals,properties,description>II)<ObjectReference,metaObject,serviceID,objectID>"
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

func fuzzCap() {

	fuzzer := gofuzz.New().NilChance(0).Funcs(makeValue).NumElements(1, 100)
	endpoint, err := net.DialEndPoint(serverURL)
	if err != nil {
		log.Fatalf("failed to contact %s: %s", serverURL, err)
	}
	defer endpoint.Close()
	perm := makeCap(fuzzer)

	file, _ := ioutil.TempFile(".", "cap-*.bin")
	err = bus.WriteCapabilityMap(perm, file)
	if err != nil {
		panic(err)
	}
	file.Close()
}

func main() {

	fuzz.WriteCorpus()

	for i := 0; i < 20; i++ {
		fuzzCap()
	}
}

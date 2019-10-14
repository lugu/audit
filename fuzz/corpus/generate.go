//go:generate go run .

package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"strings"

	gofuzz "github.com/google/gofuzz"
	"github.com/lugu/audit/fuzz"
	"github.com/lugu/qiloop/bus"
	dir "github.com/lugu/qiloop/bus/directory"
	"github.com/lugu/qiloop/bus/net"
	"github.com/lugu/qiloop/bus/session/token"
	"github.com/lugu/qiloop/bus/util"
	"github.com/lugu/qiloop/meta/signature"
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

func cleanName(c gofuzz.Continue) string {
	var name string
	c.Fuzz(&name)
	name = signature.CleanName(name)
	if name == "" {
		return cleanName(c)
	}
	return name
}

func makeStruct(i *value.Value, c gofuzz.Continue) {
	var sig strings.Builder
	var buf bytes.Buffer

	sig.WriteString("(")
	size := c.Intn(5)
	for i := 0; i < size; i++ {
		var val value.Value
		makeValue(&val, c)
		sig.WriteString(val.Signature())
		val.Write(&buf)
	}
	sig.WriteString(")<")
	sig.WriteString(cleanName(c))
	for i := 0; i < size; i++ {
		sig.WriteString(",")
		sig.WriteString(cleanName(c))
	}
	sig.WriteString(">")
	*i = value.Opaque(sig.String(), buf.Bytes())
}

func makeValue(i *value.Value, c gofuzz.Continue) {
	switch c.Intn(13) {
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
		var ref object.ObjectReference
		c.Fuzz(&ref)
		var buf bytes.Buffer
		object.WriteObjectReference(ref, &buf)
		*i = value.Opaque("o", buf.Bytes())
	case 12:
		makeStruct(i, c)
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

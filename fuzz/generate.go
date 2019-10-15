//go:generate go run ./gen -d ./corpus

package fuzz

import (
	"bytes"
	"regexp"
	"strings"

	gofuzz "github.com/google/gofuzz"
	"github.com/lugu/qiloop/bus"
	"github.com/lugu/qiloop/type/object"
	"github.com/lugu/qiloop/type/value"
)

var (
	fuzzer = gofuzz.New().NilChance(0).Funcs(makeValue).NumElements(1, 100)
)

func cleanName(c gofuzz.Continue) string {
	var name string
	c.Fuzz(&name)
	exp := regexp.MustCompile(`[^a-zA-Z][^_a-zA-Z0-9]*`)
	name = exp.ReplaceAllString(name, "")
	if name == "" {
		return cleanName(c)
	}
	return name
}

func makeEmbeddedValue(i *value.Value, c gofuzz.Continue) {
	var v value.Value
	makeValue(&v, c)
	var buf bytes.Buffer
	v.Write(&buf)
	*i = value.Opaque("m", buf.Bytes())
}

func makeObject(i *value.Value, c gofuzz.Continue) {
	var ref object.ObjectReference
	c.Fuzz(&ref)
	var buf bytes.Buffer
	object.WriteObjectReference(ref, &buf)
	*i = value.Opaque("o", buf.Bytes())
}

func makeList(i *value.Value, c gofuzz.Continue) {
	var list value.ListValue
	size := c.Intn(7)
	list = make([]value.Value, size)
	for i := 0; i < size; i++ {
		makeValue(&list[i], c)
	}
	*i = list
}

func makeStruct(i *value.Value, c gofuzz.Continue) {
	var sig strings.Builder
	var data bytes.Buffer

	sig.WriteString("(")
	size := c.Intn(7)
	for i := 0; i < size; i++ {
		var val value.Value
		makeValue(&val, c)
		sig.WriteString(val.Signature())
		data.Write(value.Bytes(val))
	}
	sig.WriteString(")<")
	sig.WriteString(cleanName(c))
	for i := 0; i < size; i++ {
		sig.WriteString(",")
		sig.WriteString(cleanName(c))
	}
	sig.WriteString(">")
	*i = value.Opaque(sig.String(), data.Bytes())
}

func makeValue(i *value.Value, c gofuzz.Continue) {
	switch c.Intn(14) {
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
		*i = value.Void()
	case 10:
		makeList(i, c)
	case 11:
		makeObject(i, c)
	case 12:
		makeStruct(i, c)
	case 13:
		makeEmbeddedValue(i, c)
	}
}

func MakeCap() bus.CapabilityMap {
	var permission bus.CapabilityMap
	fuzzer.Fuzz(&permission)
	var user string
	fuzzer.Fuzz(&user)
	permission["user"] = value.String(user)
	var token string
	fuzzer.Fuzz(&token)
	permission["token"] = value.String(token)
	return permission
}

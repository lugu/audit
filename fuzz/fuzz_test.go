package fuzz_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/lugu/audit/fuzz"
	"github.com/lugu/qiloop/bus"
)

func TestFuzzOK(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Join("testdata",
		"cap-auth-failure.bin"))
	if err != nil {
		t.Errorf("cannot open test data %s", err)
	}
	if fuzz.Fuzz(data) != 1 {
		t.Errorf("shall return 1: %d", fuzz.Fuzz(data))
	}
}

func WriteReadTest(cm bus.CapabilityMap) error {
	var buf bytes.Buffer
	err := bus.WriteCapabilityMap(cm, &buf)
	if err != nil {
		return err
	}
	_, err = bus.ReadCapabilityMap(&buf)
	if err != nil {
		return err
	}
	return nil
}

func TestSamples(t *testing.T) {
	for name, metacap := range fuzz.GetSamples() {
		err := WriteReadTest(metacap)
		if err != nil {
			t.Errorf("failed on %s: %s", name, err)
		}
	}
}

func TestCaps(t *testing.T) {
	for i := 0; i < 20; i++ {
		perm := fuzz.MakeCap()
		var buf bytes.Buffer
		err := bus.WriteCapabilityMap(perm, &buf)
		if err != nil {
			t.Errorf("failed on wirte: %s", err)
		}
		if fuzz.Fuzz(buf.Bytes()) != 1 {
			t.Errorf("shall return 1: %d", fuzz.Fuzz(buf.Bytes()))
		}
	}
}

package fuzz_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/lugu/audit/fuzz"
	"github.com/lugu/qiloop/bus"
	"github.com/lugu/qiloop/bus/directory"
	"github.com/lugu/qiloop/bus/util"
)

func TestFuzz(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("shall panic")
		}
	}()

	passwords := map[string]string{
		"nao": "nao",
	}
	addr := util.NewUnixAddr()

	auth := bus.Dictionary(passwords)
	server, err := directory.NewServer(addr, auth)
	if err != nil {
		panic(err)
	}

	fuzz.ServerURL = addr

	data, err := ioutil.ReadFile(filepath.Join("testdata", "cap-auth-failure.bin"))
	if err != nil {
		t.Errorf("cannot open test data %s", err)
	}
	fuzz.Fuzz(data)

	server.Terminate()
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

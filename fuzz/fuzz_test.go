package fuzz_test

import (
	"github.com/lugu/audit/fuzz"
	"github.com/lugu/qiloop/bus/server"
	"github.com/lugu/qiloop/bus/server/directory"
	"github.com/lugu/qiloop/bus/util"
	"io/ioutil"
	"path/filepath"
	"testing"
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

	auth := server.Dictionary(passwords)
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

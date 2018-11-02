package fuzz_test

import (
	"github.com/lugu/audit/fuzz"
	"github.com/lugu/qiloop/bus/server"
	"io/ioutil"
	gonet "net"
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
	object := server.NewServiceAuthenticate(passwords)
	service := server.NewService(object)
	router := server.NewRouter()
	router.Add(service)
	listener, err := gonet.Listen("tcp", ":9559")
	if err != nil {
		panic(err)
	}
	server := server.NewServer2(listener, router)
	go server.Run()

	fuzz.ServerURL = "tcp://localhost:9559"

	data, err := ioutil.ReadFile(filepath.Join("testdata", "cap-auth-failure.bin"))
	if err != nil {
		t.Errorf("cannot open test data %s", err)
	}
	fuzz.Fuzz(data)

	server.Stop()
}

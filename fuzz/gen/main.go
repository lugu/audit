package main

import (
	"flag"
	"io/ioutil"

	"github.com/lugu/audit/fuzz"
	"github.com/lugu/qiloop/bus"
)

func main() {
	dir := "corpus"
	flag.StringVar(&dir, "d", dir, "output directory")
	flag.Parse()

	fuzz.WriteCorpus(dir)

	for i := 0; i < 20; i++ {
		perm := fuzz.MakeCap()
		file, _ := ioutil.TempFile(dir, "cap-*.bin")
		err := bus.WriteCapabilityMap(perm, file)
		if err != nil {
			panic(err)
		}
		file.Close()
	}
}

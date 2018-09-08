package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/butlermatt/waxmoth/msg"
)

type RemoteClient struct {
	address string
	data    chan<- *msg.Message
}

type Message struct {
	raw    []byte
	origin string
}

func main() {
	a := flag.String("a", "localhost:30005", "Comma separated list of addresses and ports to connect to for input.")
	flag.Parse()

	addresses := strings.Split(*a, ",")
	fmt.Println("Received addresses: ", addresses)

	data := make(chan *msg.Message)
	for _, addr := range addresses {
		go readRemote(&RemoteClient{addr, data})
	}

	for m := range data {
		fmt.Printf("%+v\n", m)
	}
}

func readRemote(rc *RemoteClient) {
	conn, err := net.Dial("tcp", rc.address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to %q: %v\n", rc.address, err)
		return
	}
	defer conn.Close()

	parser := make(chan *msg.Raw)
	go msg.ParseChannel(parser, rc.data)

	buf := bufio.NewReader(conn)
	for {
		data, err := buf.ReadBytes('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read bytes on %q: %v\n", rc.address, err)
			continue
		}

		rm := &msg.Raw{Origin: rc.address, Data: bytes.TrimSpace(data)}
		parser <- rm
	}
}

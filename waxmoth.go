package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

type RemoteClient struct {
	address string
	data    chan<- *Message
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

	data := make(chan *Message)
	for _, addr := range addresses {
		go readRemote(&RemoteClient{addr, data})
	}

	for line := range data {
		fmt.Printf("%s - %s\n", line.origin, line.raw)
	}
}

func readRemote(rc *RemoteClient) {
	conn, err := net.Dial("tcp", rc.address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to %q: %v\n", rc.address, err)
		return
	}
	defer conn.Close()

	buf := bufio.NewReader(conn)
	for {
		data, err := buf.ReadBytes('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read bytes on %q: %v\n", rc.address, err)
			continue
		}
		rc.data <- &Message{bytes.TrimSpace(data), rc.address}
	}
}

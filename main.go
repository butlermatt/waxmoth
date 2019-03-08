package main

import (
	"flag"
	"log"

	"github.com/butlermatt/waxmoth/server"
)

func main() {
	port := flag.String("p", "8888", "port")
	flag.Parse()
	s := server.New()
	log.Print("Starting server at: http://localhost:", *port)
	log.Fatal(s.ListenAndServe(":" + *port))
}

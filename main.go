package main

import (
	"log"

	"github.com/you/go-web-server/server"
)

func main() {
	srv := server.Server()
	log.Println("Server listening on", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
package main

import (
	"log"
	"os"

	server "gitlab.com/robrohan/knotset/internals"
)

func main() {
	if err := server.Run(); err != nil {
		log.Println("error :", err)
		os.Exit(1)
	}
}

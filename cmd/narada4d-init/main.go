package main

import (
	"log"

	_ "github.com/powerman/narada4d/protocol/file"
	_ "github.com/powerman/narada4d/protocol/goose-postgres"
	_ "github.com/powerman/narada4d/protocol/mysql"
	"github.com/powerman/narada4d/schemaver"
)

func main() {
	log.SetFlags(0)

	err := schemaver.Initialize()
	if err != nil {
		log.Fatalln("Failed to initialize:", err)
	}
}

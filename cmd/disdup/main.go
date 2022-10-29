package main

import (
	"flag"
	"log"

	"github.com/ethanv2/disdup"
)

// Flags.
var (
	AuthToken = flag.String("token", "", "Bot authentication token")
)

func main() {
	flag.Parse()

	if *AuthToken == "" {
		log.Fatalln("disdup: auth token required but not provided")
	}

	dup, err := disdup.NewDuplicator(*AuthToken)
	if err != nil {
		log.Fatalln(err)
	}
	defer dup.Close()
}

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/ethanv2/disdup"
)

// Command line flags.
var (
	AuthToken = flag.String("token", "", "Bot authentication token")
)

func main() {
	flag.Parse()
	if *AuthToken == "" {
		log.Fatalln("disdup: auth token required but not provided")
	}

	log.Println("Connecting to discord...")
	dup, err := disdup.NewDuplicator(*AuthToken)
	if err != nil {
		log.Fatalln(err)
	}
	defer dup.Close()
	log.Println("Connection to Discord established")

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select {
	case <-sigint:
		log.Println("Caught interrupt. Terminating gracefully")
	case err := <-dup.Wait():
		log.Println(err)
	}
}

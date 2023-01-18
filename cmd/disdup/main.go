package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/ejv2/disdup"
	clconf "github.com/ejv2/disdup/cmd/disdup/conf"
)

// Command line flags.
var (
	AuthToken = flag.String("token", "", "Bot authentication token")
)

func main() {
	cfg, err := clconf.LoadConfig()
	if err != nil {
		log.Fatal("config error: ", err)
	}

	if *AuthToken != "" {
		cfg.Token = *AuthToken
	}

	log.Println("Connecting to discord...")
	dup, err := disdup.NewDuplicator(cfg)
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

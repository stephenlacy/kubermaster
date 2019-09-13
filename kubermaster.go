package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/cors"
	"github.com/stevelacy/kubermaster/manager"
)

var version = "dev"

func main() {
	fmt.Printf("Starting kubermaster, version: %v \n", version)

	tokenDescription := "Authentication token for cluster" +
		"(or ENV TOKEN=<token>)"
	portDescription := "Port for manager to listen on, 9090 by default"
	memoryDescription := "Memory limit for each task, 250 by default"

	token := flag.String("token", "", tokenDescription)
	port := flag.String("port", "", portDescription)
	memory := flag.String("memory", "", memoryDescription)

	if *token == "" {
		*token = os.Getenv("TOKEN")
	}

	if *port == "" {
		*port = os.Getenv("PORT")
		if *port == "" {
			*port = "9090"
		}
	}

	if *memory == "" {
		envmem := os.Getenv("MEMORY")
		*memory = envmem
	}

	flag.Parse()

	if *token == "" {
		fmt.Println("Error: Missing required parameters")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Println("Listening on port:", *port, "With memory limit:", *memory)

	origin := os.Getenv("ORIGIN")
	if origin == "" {
		origin = "http://localhost:1234"
	}
	c := cors.New(cors.Options{
		AllowedOrigins: []string{origin},
	})
	handler := c.Handler(manager.Init(*token, *memory))

	err := http.ListenAndServe(fmt.Sprintf(":%v", *port), handler)
	if err != nil {
		panic(err)
	}
}

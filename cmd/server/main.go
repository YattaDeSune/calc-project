package main

import (
	"log"

	"github.com/YattaDeSune/calc-project/internal/server"
)

func main() {
	server := server.New()
	if err := server.RunServer(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

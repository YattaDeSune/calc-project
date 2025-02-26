package main

import (
	"log"

	"github.com/YattaDeSune/calc-project/internal/application/agent"
)

func main() {
	agent := agent.New()
	if err := agent.RunAgent(); err != nil {
		log.Fatalf("Failed to run agent: %v", err)
	}
}

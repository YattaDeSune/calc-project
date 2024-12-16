package main

import (
	"log"
	"os"

	"github.com/YattaDeSune/calc-project/internal/application"
)

func main() {
	app := application.New()
	if err := app.RunServer(); err != nil {
		log.Fatalf("[ERROR] Failed to start server: %v", err)
		os.Exit(1)
	}
}

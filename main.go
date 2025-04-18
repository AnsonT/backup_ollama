package main

import (
	"log"
	"os"

	"backup_ollama/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

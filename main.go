package main

import (
	"fmt"
	"os"
)

func printUsage() {
	fmt.Println(`ollama-mem — persistent memory for Ollama

Usage:
  ollama-mem remember "text to store"
  ollama-mem ask "your question"
  ollama-mem list
  ollama-mem clear`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
}

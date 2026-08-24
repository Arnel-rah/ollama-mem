package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const memoryRelPath = ".ollama-mem/memory.jsonl"

type Memory struct {
	Text      string    `json:"text"`
	Embedding []float64 `json:"embedding"`
	CreatedAt time.Time `json:"created_at"`
}

func memoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, memoryRelPath)
}

func loadMemories() ([]Memory, error) {
	path := memoryPath()
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return []Memory{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var mems []Memory
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var m Memory
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		mems = append(mems, m)
	}
	return mems, scanner.Err()
}

func appendMemory(m Memory) error {
	path := memoryPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line, err := json.Marshal(m)
	if err != nil {
		return err
	}
	_, err = f.Write(append(line, '\n'))
	return err
}

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

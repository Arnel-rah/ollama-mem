package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ollamaURL     = "http://localhost:11434"
	embedModel    = "nomic-embed-text"
	memoryRelPath = ".ollama-mem/memory.jsonl"
)

type Memory struct {
	Text      string    `json:"text"`
	Embedding []float64 `json:"embedding"`
	CreatedAt time.Time `json:"created_at"`
}

type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

func memoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, memoryRelPath)
}

func embed(text string) ([]float64, error) {
	body, _ := json.Marshal(embedRequest{Model: embedModel, Input: text})
	resp, err := http.Post(ollamaURL+"/api/embed", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embed call: %w", err)
	}
	defer resp.Body.Close()

	var er embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		return nil, fmt.Errorf("embed decode: %w", err)
	}
	if len(er.Embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned, is model '%s' pulled?", embedModel)
	}
	return er.Embeddings[0], nil
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

func remember(text string) error {
	vec, err := embed(text)
	if err != nil {
		return err
	}
	m := Memory{Text: text, Embedding: vec, CreatedAt: time.Now()}
	if err := appendMemory(m); err != nil {
		return err
	}
	fmt.Printf("Remembered: %q\n", text)
	return nil
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

	cmd := os.Args[1]
	args := strings.Join(os.Args[2:], " ")

	var err error
	switch cmd {
	case "remember":
		if args == "" {
			fmt.Println("Error: provide text to remember")
			os.Exit(1)
		}
		err = remember(args)
	default:
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

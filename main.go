package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	ollamaURL     = "http://localhost:11434"
	embedModel    = "nomic-embed-text"
	chatModel     = "qwen2.5-coder:7b"
	topK          = 5
	memoryRelPath = ".ollama-mem/memory.jsonl"
)

var version = "dev"

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

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatResponseChunk struct {
	Message chatMessage `json:"message"`
	Done    bool        `json:"done"`
}

func memoryPath() string {
	if dir := os.Getenv("OLLAMA_MEM_DIR"); dir != "" {
		return filepath.Join(dir, "memory.jsonl")
	}
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
		return nil, fmt.Errorf("appel embed: %w", err)
	}
	defer resp.Body.Close()

	var er embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		return nil, fmt.Errorf("decode embed: %w", err)
	}
	if len(er.Embeddings) == 0 {
		return nil, fmt.Errorf("aucun embedding retourné, modèle '%s' disponible ?", embedModel)
	}
	return er.Embeddings[0], nil
}

func cosineSim(a, b []float64) float64 {
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
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

type scored struct {
	mem   Memory
	score float64
}

const minSimilarityScore = 0.3

func rankMemories(qvec []float64, mems []Memory) []scored {
	scoredMems := make([]scored, 0, len(mems))
	for _, m := range mems {
		scoredMems = append(scoredMems, scored{mem: m, score: cosineSim(qvec, m.Embedding)})
	}
	sort.Slice(scoredMems, func(i, j int) bool { return scoredMems[i].score > scoredMems[j].score })
	return scoredMems
}

func formatContext(ranked []scored) string {
	n := topK
	if n > len(ranked) {
		n = len(ranked)
	}

	var sb strings.Builder
	for i := 0; i < n; i++ {
		if ranked[i].score < minSimilarityScore {
			continue
		}
		sb.WriteString("- ")
		sb.WriteString(ranked[i].mem.Text)
		sb.WriteString("\n")
	}
	return sb.String()
}

func retrieveContext(query string) (string, error) {
	mems, err := loadMemories()
	if err != nil {
		return "", err
	}
	if len(mems) == 0 {
		return "", nil
	}

	qvec, err := embed(query)
	if err != nil {
		return "", err
	}

	ranked := rankMemories(qvec, mems)
	return formatContext(ranked), nil
}

func chatWithMemory(query string) error {
	context, err := retrieveContext(query)
	if err != nil {
		return err
	}

	systemPrompt := "Tu es un assistant technique utile."
	if context != "" {
		systemPrompt += "\n\nContexte utile mémorisé sur l'utilisateur :\n" + context
	}

	reqBody := chatRequest{
		Model: chatModel,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: query},
		},
		Stream: true,
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(ollamaURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("appel chat: %w", err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk chatResponseChunk
		if err := decoder.Decode(&chunk); err != nil {
			break
		}
		fmt.Print(chunk.Message.Content)
		if chunk.Done {
			break
		}
	}
	fmt.Println()
	return nil
}

func printUsage() {
	fmt.Println(`ollama-mem — mémoire persistante pour Ollama

Usage:
  ollama-mem remember "texte à retenir"
  ollama-mem ask "ta question"
  ollama-mem list
  ollama-mem clear`)
}

func listMemories() error {
	mems, err := loadMemories()
	if err != nil {
		return err
	}
	if len(mems) == 0 {
		fmt.Println("No memories stored.")
		return nil
	}
	for i, m := range mems {
		fmt.Printf("%d. [%s] %s\n", i+1, m.CreatedAt.Format("2006-01-02 15:04"), m.Text)
	}
	return nil
}

func clearMemories() error {
	path := memoryPath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	fmt.Println("Memory cleared.")
	return nil
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
	case "ask":
		if args == "" {
			fmt.Println("Error: provide a question")
			os.Exit(1)
		}
		err = chatWithMemory(args)
	case "list":
		err = listMemories()
	case "clear":
		err = clearMemories()
	case "version":
		fmt.Println(version)
	default:
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

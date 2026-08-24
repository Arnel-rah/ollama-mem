package main

import "testing"

func TestRankMemoriesOrdersBySimilarityDescending(t *testing.T) {
	query := []float64{1, 0}
	mems := []Memory{
		{Text: "far", Embedding: []float64{0, 1}},
		{Text: "close", Embedding: []float64{1, 0.01}},
		{Text: "medium", Embedding: []float64{1, 1}},
	}

	ranked := rankMemories(query, mems)

	if ranked[0].mem.Text != "close" {
		t.Errorf("expected 'close' ranked first, got %q", ranked[0].mem.Text)
	}
	if ranked[len(ranked)-1].mem.Text != "far" {
		t.Errorf("expected 'far' ranked last, got %q", ranked[len(ranked)-1].mem.Text)
	}
}

func TestFormatContextFiltersLowScores(t *testing.T) {
	ranked := []scored{
		{mem: Memory{Text: "relevant"}, score: 0.8},
		{mem: Memory{Text: "irrelevant"}, score: 0.1},
	}

	got := formatContext(ranked)

	if got != "- relevant\n" {
		t.Errorf("formatContext = %q, want only the entry above threshold", got)
	}
}

func TestFormatContextRespectsTopK(t *testing.T) {
	ranked := make([]scored, 0, topK+3)
	for i := 0; i < topK+3; i++ {
		ranked = append(ranked, scored{mem: Memory{Text: "entry"}, score: 0.9})
	}

	got := formatContext(ranked)
	lines := 0
	for _, r := range []rune(got) {
		if r == '\n' {
			lines++
		}
	}

	if lines != topK {
		t.Errorf("formatContext produced %d lines, want %d (topK)", lines, topK)
	}
}

func TestFormatContextEmptyWhenAllBelowThreshold(t *testing.T) {
	ranked := []scored{
		{mem: Memory{Text: "weak match"}, score: 0.1},
	}

	got := formatContext(ranked)

	if got != "" {
		t.Errorf("formatContext = %q, want empty string", got)
	}
}

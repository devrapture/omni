package service

import (
	"strings"
	"testing"
)

func TestChunkTextSplitsSingleLongSentence(t *testing.T) {
	cfg := ChunkConfig{ChunkSize: 25, Overlap: 5}
	chunks := ChunkText(words(73), cfg)

	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d: %#v", len(chunks), chunks)
	}
	assertChunksAtMost(t, chunks, cfg.ChunkSize)
}

func TestChunkTextDoesNotLetOverlapCreateOversizedChunk(t *testing.T) {
	cfg := ChunkConfig{ChunkSize: 30, Overlap: 10}
	text := words(25) + ". " + words(25) + "."

	chunks := ChunkText(text, cfg)

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d: %#v", len(chunks), chunks)
	}
	assertChunksAtMost(t, chunks, cfg.ChunkSize)
}

func assertChunksAtMost(t *testing.T, chunks []string, maxWords int) {
	t.Helper()
	for i, chunk := range chunks {
		if got := countWords(chunk); got > maxWords {
			t.Fatalf("chunk %d has %d words, want at most %d: %q", i, got, maxWords, chunk)
		}
	}
}

func words(count int) string {
	values := make([]string, count)
	for i := range values {
		values[i] = "word"
	}
	return strings.Join(values, " ")
}

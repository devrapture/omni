package service

import (
	"strings"
	"unicode"
)

// ChunkConfig controls how text is split into chunks.
type ChunkConfig struct {
	// ChunkSize is the target number of words per chunk.
	// 300-500 words is the sweet spot: enough context, small enough to be specific.
	ChunkSize int

	// Overlap is how many words are shared between adjacent chunks.
	// Overlap prevents answers from falling at chunk boundaries.
	// 50-100 words is typical.
	Overlap int
}

// DefaultChunkConfig returns sensible defaults for business knowledge.
func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		ChunkSize: 400,
		Overlap:   80,
	}
}

// ChunkText splits text into overlapping chunks.
//
// Strategy:
// 1. Split text into sentences (on ., ?, !)
// 2. Group sentences until we hit ChunkSize words
// 3. Start the next chunk Overlap words before the end of the current chunk
//
// This preserves sentence boundaries, which improves readability and
// semantic coherence compared to splitting on raw word count.
func ChunkText(text string, cfg ChunkConfig) []string {
	// Clean the text
	text = strings.TrimSpace(text)
	text = normalizeWhitespace(text)

	if text == "" {
		return nil
	}

	// Split into sentences
	sentences := splitIntoSentences(text)
	sentences = splitLongSentences(sentences, cfg.ChunkSize)

	if len(sentences) == 0 {
		return nil
	}

	var chunks []string
	var currentChunk []string
	currentWordCount := 0

	for _, sentence := range sentences {
		wordCount := countWords(sentence)

		// If adding this sentence exceeds our chunk size AND we have content,
		// save the current chunk and start a new one with overlap
		if currentWordCount+wordCount > cfg.ChunkSize && currentWordCount > 0 {
			chunk := strings.Join(currentChunk, " ")
			chunks = append(chunks, strings.TrimSpace(chunk))

			// Create overlap only when it fits with the next sentence.
			overlapWords := cfg.Overlap
			if remaining := cfg.ChunkSize - wordCount; remaining < overlapWords {
				overlapWords = remaining
			}
			currentChunk = getOverlapSentences(currentChunk, overlapWords)
			currentWordCount = countWords(strings.Join(currentChunk, " "))
		}

		currentChunk = append(currentChunk, sentence)
		currentWordCount += wordCount
	}

	// Don't forget the last chunk
	if len(currentChunk) > 0 {
		chunk := strings.Join(currentChunk, " ")
		if strings.TrimSpace(chunk) != "" {
			chunks = append(chunks, strings.TrimSpace(chunk))
		}
	}

	// Filter out very short chunks (less than 20 words)
	// These are usually headers or noise
	var filtered []string
	for _, c := range chunks {
		if countWords(c) >= 20 {
			filtered = append(filtered, c)
		}
	}

	return filtered
}

func splitLongSentences(sentences []string, maxWords int) []string {
	if maxWords <= 0 {
		return sentences
	}

	var result []string
	for _, sentence := range sentences {
		if countWords(sentence) <= maxWords {
			result = append(result, sentence)
			continue
		}
		result = append(result, splitIntoWordChunks(sentence, maxWords)...)
	}
	return result
}

func splitIntoWordChunks(text string, maxWords int) []string {
	words := strings.Fields(text)
	if maxWords <= 0 || len(words) <= maxWords {
		return []string{strings.TrimSpace(text)}
	}

	chunks := make([]string, 0, (len(words)+maxWords-1)/maxWords)
	for i := 0; i < len(words); i += maxWords {
		end := i + maxWords
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[i:end], " "))
	}
	return chunks
}

// splitIntoSentences splits text on sentence-ending punctuation.
// It's not perfect (e.g., "Dr. Smith" would split), but it's fast and
// good enough for business content.
func splitIntoSentences(text string) []string {
	var sentences []string
	var current strings.Builder

	runes := []rune(text)
	for i, r := range runes {
		current.WriteRune(r)

		// Sentence ends at . ? ! followed by whitespace or end of text
		if r == '.' || r == '?' || r == '!' || r == '\n' {
			next := i + 1
			if next >= len(runes) || unicode.IsSpace(runes[next]) {
				sentence := strings.TrimSpace(current.String())
				if sentence != "" {
					sentences = append(sentences, sentence)
				}
				current.Reset()
			}
		}
	}

	// Capture any remaining text that doesn't end with punctuation
	if remaining := strings.TrimSpace(current.String()); remaining != "" {
		sentences = append(sentences, remaining)
	}

	return sentences
}

// getOverlapSentences returns the last sentences that together have at most `targetWords` words.
func getOverlapSentences(sentences []string, targetWords int) []string {
	if len(sentences) == 0 || targetWords <= 0 {
		return nil
	}

	var overlap []string
	wordCount := 0

	// Walk backwards through sentences
	for i := len(sentences) - 1; i >= 0; i-- {
		wc := countWords(sentences[i])
		if wordCount+wc > targetWords {
			break
		}
		overlap = append([]string{sentences[i]}, overlap...)
		wordCount += wc
	}

	return overlap
}

// countWords counts whitespace-separated tokens.
func countWords(s string) int {
	return len(strings.Fields(s))
}

// normalizeWhitespace collapses multiple spaces/newlines into single spaces.
func normalizeWhitespace(s string) string {
	// Replace multiple whitespace characters with a single space
	var result strings.Builder
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				result.WriteRune(' ')
			}
			prevSpace = true
		} else {
			result.WriteRune(r)
			prevSpace = false
		}
	}
	return result.String()
}

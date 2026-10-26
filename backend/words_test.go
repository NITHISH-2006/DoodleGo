package main

import (
	"regexp"
	"testing"
)

func TestDifficultyWordBanks(t *testing.T) {
	banks := map[string][]string{
		difficultyEasy:   easyWords,
		difficultyMedium: mediumWords,
		difficultyHard:   hardWords,
	}
	validWord := regexp.MustCompile(`^[a-z]+$`)
	seen := make(map[string]string)

	for difficulty, words := range banks {
		if len(words) < 200 {
			t.Fatalf("%s bank has only %d words, want at least 200", difficulty, len(words))
		}
		within := make(map[string]bool)
		for _, word := range words {
			if !validWord.MatchString(word) {
				t.Fatalf("%s word %q must contain lowercase letters only", difficulty, word)
			}
			if within[word] {
				t.Fatalf("%s bank contains duplicate word %q", difficulty, word)
			}
			within[word] = true
			if other, exists := seen[word]; exists {
				t.Fatalf("word %q appears in both %s and %s", word, other, difficulty)
			}
			seen[word] = difficulty
		}
	}

	if len(seen) < 600 {
		t.Fatalf("expected at least 600 unique words, got %d", len(seen))
	}
}

func TestWordDeckCoversBankWithoutRepeats(t *testing.T) {
	room := testRoom(modeClassic, 0)
	room.Difficulty = difficultyHard
	used := make(map[string]bool)
	for i := 0; i < len(hardWords); i++ {
		word := room.pickWordLocked()
		if used[word] {
			t.Fatalf("word %q repeated before the deck was exhausted", word)
		}
		used[word] = true
	}
	if len(used) != len(hardWords) {
		t.Fatalf("expected %d unique words, got %d", len(hardWords), len(used))
	}
}

func TestWordDeckRefillAvoidsImmediateRepeat(t *testing.T) {
	room := testRoom(modeClassic, 0)
	room.Difficulty = difficultyEasy
	room.lastWord = easyWords[0]
	previous := room.lastWord
	room.wordDeck = nil
	room.wordDeckIndex = 0

	word := room.pickWordLocked()
	if word == previous && len(easyWords) > 1 {
		t.Fatalf("refilled deck immediately repeated %q", word)
	}
}

func TestNormalizeDifficultyDefaultsToEasy(t *testing.T) {
	cases := map[string]string{
		"":        difficultyEasy,
		"EASY":    difficultyEasy,
		"medium":  difficultyMedium,
		" Hard ":  difficultyHard,
		"unknown": difficultyEasy,
	}
	for input, want := range cases {
		if got := normalizeDifficulty(input); got != want {
			t.Fatalf("normalizeDifficulty(%q) = %q, want %q", input, got, want)
		}
	}
}

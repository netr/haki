package anki_test

import (
	"testing"

	"github.com/netr/haki/anki"
)

func Test_RemoveParentDecks(t *testing.T) {
	t.Parallel()

	test := []struct {
		name      string
		deckNames anki.DeckNames
		expected  []string
	}{
		{
			name:      "all parent decks removed, return empty slice",
			deckNames: anki.DeckNames{"deck1", "deck2", "deck3"},
			expected:  []string{},
		},
		{
			name:      "some parent decks removed, return slice with parent decks",
			deckNames: anki.DeckNames{"deck1", "deck2::subdeck", "deck3::subdeck::subsubdeck"},
			expected:  []string{"deck2::subdeck", "deck3::subdeck::subsubdeck"},
		},
		{
			name:      "no parent decks removed, return slice with all decks",
			deckNames: anki.DeckNames{"deck1::subdeck", "deck2::subdeck", "deck3::subdeck"},
			expected:  []string{"deck1::subdeck", "deck2::subdeck", "deck3::subdeck"},
		},
	}

	for _, tc := range test {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := anki.RemoveParentDecks(tc.deckNames)
			if len(actual) != len(tc.expected) {
				t.Errorf("expected %d decks, got %d", len(tc.expected), len(actual))
			}

			for i, deck := range actual {
				if deck != tc.expected[i] {
					t.Errorf("expected deck %s, got %s", tc.expected[i], deck)
				}
			}
		})
	}
}

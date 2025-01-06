package anki_test

import (
	"slices"
	"testing"

	"github.com/netr/haki/anki"
)

func Test_FilterDecksByHierarchy(t *testing.T) {
	t.Parallel()

	test := []struct {
		name      string
		deckNames anki.DeckNames
		expected  []string
	}{
		{
			name:      "all solo decks",
			deckNames: anki.DeckNames{"deck1", "deck2", "deck3"},
			expected:  []string{"deck3", "deck2", "deck1"},
		},
		{
			name:      "parent deck should be removed and only show child",
			deckNames: anki.DeckNames{"deck1", "deck1::test"},
			expected:  []string{"deck1::test"},
		},
		{
			name:      "some parent decks removed, return slice with parent decks",
			deckNames: anki.DeckNames{"deck1", "deck2::subdeck", "deck3::subdeck::subsubdeck"},
			expected:  []string{"deck2::subdeck", "deck3::subdeck::subsubdeck", "deck1"},
		},
		{
			name:      "all parent decks removed, return slice with all decks",
			deckNames: anki.DeckNames{"deck1", "deck1::subdeck", "deck2", "deck2::subdeck", "deck3", "deck3::subdeck"},
			expected:  []string{"deck1::subdeck", "deck2::subdeck", "deck3::subdeck"},
		},
		{
			name:      "re-arranged decks still work properly",
			deckNames: anki.DeckNames{"deck1", "deck2::subdeck", "deck2", "deck1::subdeck"},
			expected:  []string{"deck1::subdeck", "deck2::subdeck"},
		},
	}

	for _, tc := range test {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := anki.FilterDecksByHierarchy(tc.deckNames)
			if len(actual) != len(tc.expected) {
				t.Errorf("expected %d decks, got %d", len(tc.expected), len(actual))
			}

			for _, deck := range actual {
				if !slices.Contains(tc.expected, deck) {
					t.Errorf("expected deck: %s", deck)
				}
			}
		})
	}
}

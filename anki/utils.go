package anki

import "strings"

func RemoveParentDecks(deckNames DeckNames) []string {
	filtered := []string{}
	for _, deck := range deckNames {
		if strings.Contains(deck, "::") {
			filtered = append(filtered, deck)
		}
	}
	return filtered
}

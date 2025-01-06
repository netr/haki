package anki

import (
	"slices"
	"strings"
)

// FilterDecksByHierarchy returns decks that are either:
// - Solo decks (no parent/child relationship)
// - Child decks (contain :: separator)
// The function filters out parent decks that have children.
func FilterDecksByHierarchy(deckNames DeckNames) []string {
	// This needs to be sorted to ensure the parent's are found before the children.
	// Effectively allowing dev to pass in any ordering.
	slices.Sort(deckNames)

	result := make([]string, 0, len(deckNames))
	parentHasChildren := make(map[string]bool, len(deckNames))

	for _, deckName := range deckNames {
		if strings.Contains(deckName, "::") {
			parentName := strings.Split(deckName, "::")[0]
			parentHasChildren[parentName] = true
			result = append(result, deckName)
			continue
		}

		// Mark potential solo deck as not having children
		parentHasChildren[deckName] = false
	}

	// Add back solo decks (parents without children)
	for deckName, hasChildren := range parentHasChildren {
		if !hasChildren {
			result = append(result, deckName)
		}
	}

	return result
}

package anki

import "fmt"

type DeckNameService struct {
	client *Client
}

func NewDeckNameService(client *Client) *DeckNameService {
	return &DeckNameService{client}
}

// DeckNames is a slice of deck names.
type DeckNames []string

// GetNames returns a slice of deck names.
func (svc *DeckNameService) GetNames() (DeckNames, error) {
	var decks DeckNames
	if err := svc.client.sendAndUnmarshal("deckNames", nil, &decks); err != nil {
		return nil, fmt.Errorf("deckNames: %w", err)
	}
	return decks, nil
}

// DeckNamesAndIds is a map of deck names and their corresponding IDs.
type DeckNamesAndIds map[string]float64

// GetNamesAndIds returns a map of deck names and their corresponding IDs.
func (svc *DeckNameService) GetNamesAndIds() (DeckNamesAndIds, error) {
	var decks DeckNamesAndIds
	if err := svc.client.sendAndUnmarshal("deckNamesAndIds", nil, &decks); err != nil {
		return nil, fmt.Errorf("deckNamesAndIds: %w", err)
	}
	return decks, nil
}

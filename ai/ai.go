// Package ai provides interfaces and types for AI-powered card creation.

package ai

import "errors"

// Error variables for common error cases.
var (
	ErrUnimplemented          = errors.New("unimplemented")
	ErrInvalidAIModelProvider = errors.New("invalid ai model provider")
)

// CardCreator defines the interface for creating Anki cards.
type CardCreator interface {
	// ChooseDeck selects a deck based on provided deck names and text.
	ChooseDeck(deckNames []string, text string) (string, error)
	// CreateAnkiCards generates Anki cards for the given deck and text.
	CreateAnkiCards(deckName string, text string) ([]AnkiCard, error)
}

// Modeler represents an AI model with a string representation.
type Modeler interface {
	String() string
}

// APIProvider defines the interface for AI API providers.
type APIProvider interface {
	Action() CardCreator
	ModelName() Modeler
}

// APIProviderName represents the name of an AI API provider.
type APIProviderName string

// Supported API providers.
const (
	OpenAI    APIProviderName = "openai"
	Anthropic APIProviderName = "anthropic"
)

// NewAPIProvider creates a new AIAPIProvider based on the given name and API key.
// It optionally accepts a Modeler to specify the model type.
func NewAPIProvider(name APIProviderName, apiKey string, modelType ...Modeler) (APIProvider, error) {
	switch name {
	case OpenAI:
		if len(modelType) == 0 {
			return &OpenAIAPIProvider{
				model: NewOpenAIModel(apiKey),
			}, nil
		}

		mt := modelType[0]
		if !isValidOpenAIModelName(mt.String()) {
			return nil, ErrInvalidOpenAIModel
		}
		return &OpenAIAPIProvider{
			model: NewOpenAIModel(apiKey, OpenAIModelName(mt.String())),
		}, nil
	case Anthropic:
		return nil, ErrUnimplemented
	default:
		return nil, ErrInvalidAIModelProvider
	}
}

// AnkiCard represents a single Anki flashcard with a front and back side.
type AnkiCard struct {
	Front string
	Back  string
}

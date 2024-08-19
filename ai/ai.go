// Package ai provides interfaces and types for AI-powered card creation.

package ai

import "errors"

// Error variables for common error cases.
var (
	ErrUnimplemented          = errors.New("unimplemented")
	ErrInvalidAPIProviderName = errors.New("invalid ai api provider name")
)

// APIProviderName represents the name of an AI API provider.
type APIProviderName string

// Supported API providers.
const (
	OpenAI    APIProviderName = "openai"
	Anthropic APIProviderName = "anthropic"
)

// Modeler represents an AI model with a string representation.
type Modeler interface {
	String() string
}

// AICardCreator defines the interface for AI API providers.
type AICardCreator interface {
	// ChooseDeck selects a deck based on provided deck names and text.
	ChooseDeck(deckNames []string, text string) (string, error)
	// Create generates Anki cards for the given deck and text.
	Create(deckName string, text string) ([]AnkiCard, error)
	// ModelName returns the model name used by the AI API provider.
	ModelName() Modeler
}

// NewAICardCreator creates a new AICardCreator based on the given name and API key.
// It optionally accepts a Modeler to specify the model type.
func NewAICardCreator(name APIProviderName, apiKey string, modelType ...Modeler) (AICardCreator, error) {
	switch name {
	case OpenAI:
		if len(modelType) == 0 {
			return NewOpenAICardCreator(apiKey)
		}

		mt := modelType[0]
		if !isValidOpenAIModelName(mt.String()) {
			return nil, ErrInvalidOpenAIModel
		}
		return NewOpenAICardCreator(apiKey, OpenAIModelName(mt.String()))
	case Anthropic:
		return nil, ErrUnimplemented
	default:
		return nil, ErrInvalidAPIProviderName
	}
}

// AnkiCard represents a single Anki flashcard with a front and back side.
type AnkiCard struct {
	Front string
	Back  string
}

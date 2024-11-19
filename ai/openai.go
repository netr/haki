package ai

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var (
	ErrInvalidOpenAIModel = errors.New("invalid openai model")
)

type ErrMissingKey struct {
	Key string
}

func (e ErrMissingKey) Error() string {
	return "missing key: " + e.Key
}

// OpenAIClient wraps the OpenAI API client
type OpenAIClient struct {
	client    *openai.Client
	modelType OpenAIModelName
}

// NewOpenAIClient creates a new OpenAI API client with the given API key and an optional model type.
// If no model type is provided, the default model is GPT4o20240806, which supports structured outputs.
func NewOpenAIClient(apiKey string, modelType ...OpenAIModelName) *OpenAIClient {
	mt := GPT4o20240806
	if len(modelType) > 0 {
		mt = modelType[0]
	}

	client := openai.NewClient(apiKey)

	return &OpenAIClient{
		modelType: mt,
		client:    client,
	}
}

// createChatCompletion allows us to avoid having to call s.client.client.CreateChatCompletion.
func (api *OpenAIClient) createChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return api.client.CreateChatCompletion(ctx, request)
}

// OpenAICardCreator is an implementation of the AICardCreator interface for OpenAI.
type OpenAICardCreator struct {
	client *OpenAIClient
}

// NewOpenAICardCreator creates a new OpenAICardCreator with the given API key and an optional model type.
func NewOpenAICardCreator(apiKey string, modelType ...OpenAIModelName) (*OpenAICardCreator, error) {
	mt := GPT4o20240806
	if len(modelType) > 0 {
		mt = modelType[0]
	}

	if !isValidOpenAIModelName(string(mt)) {
		return nil, ErrInvalidOpenAIModel
	}

	client := NewOpenAIClient(apiKey, mt)
	return &OpenAICardCreator{client: client}, nil
}

// ModelName returns the model name used by OpenAI.
func (s *OpenAICardCreator) ModelName() ModelNamer {
	return s.client.modelType
}

// ChooseDeck uses the OpenAI API to select a deck based on provided deck names and text.
func (s *OpenAICardCreator) ChooseDeck(ctx context.Context, deckNames []string, text string) (string, error) {
	deckNameChoices := strings.Join(deckNames, ", ")

	resp, err := s.client.createChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: s.ModelName().String(),
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: "Please enter a string, and we will select the best Anki card deck to place it in: " +
						"Deck: Pick the most reasonable choice from the following ```" + deckNameChoices + "```",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			},
			Tools: []openai.Tool{
				{
					Type: openai.ToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name:   "deck_selection",
						Strict: true,
						Parameters: &jsonschema.Definition{
							Type: jsonschema.Object,
							Properties: map[string]jsonschema.Definition{
								"Deck": {
									Type: jsonschema.String,
								},
							},
							Required:             []string{"Deck"},
							AdditionalProperties: false,
						},
					},
				},
			},
			ToolChoice: openai.ToolChoice{
				Type: openai.ToolTypeFunction,
				Function: openai.ToolFunction{
					Name: "deck_selection",
				},
			},
		},
	)
	if err != nil {
		return "", err
	}

	var result = make(map[string]string)
	err = json.Unmarshal([]byte(resp.Choices[0].Message.ToolCalls[0].Function.Arguments), &result)
	if err != nil {
		return "", err
	}
	for _, key := range []string{"Deck"} {
		if _, ok := result[key]; !ok {
			return "", ErrMissingKey{Key: key}
		}
	}

	return result["Deck"], nil
}

// Create uses the OpenAI API to generate AnkiCard's (front and back) for the given deck and text.
func (s *OpenAICardCreator) Create(ctx context.Context, deckName string, text string, prompt string) ([]AnkiCard, error) {

	resp, err := s.client.createChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:     s.ModelName().String(),
			MaxTokens: 4096,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			},
			Tools: []openai.Tool{
				{
					Type: openai.ToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name:   "anki_card_creation",
						Strict: true,
						Parameters: &jsonschema.Definition{
							Type: jsonschema.Object,
							Properties: map[string]jsonschema.Definition{
								"cards": {
									Type: jsonschema.Array,
									Items: &jsonschema.Definition{
										Type: jsonschema.Object,
										Properties: map[string]jsonschema.Definition{
											"front": {
												Type:        jsonschema.String,
												Description: "The front side of the card. Example: 'What is the capital of France?'",
											},
											"back": {
												Type:        jsonschema.String,
												Description: "The back side of the card. Example: 'Paris'",
											},
										},
										AdditionalProperties: false,
										Required:             []string{"front", "back"},
									},
									AdditionalProperties: false,
								},
							},
							Required:             []string{"cards"},
							AdditionalProperties: false,
						},
					},
				},
			},
			ToolChoice: openai.ToolChoice{
				Type: openai.ToolTypeFunction,
				Function: openai.ToolFunction{
					Name: "anki_card_creation",
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	var data createAnkiCardsData
	err = json.Unmarshal([]byte(resp.Choices[0].Message.ToolCalls[0].Function.Arguments), &data)
	if err != nil {
		return nil, err
	}

	return data.Cards, nil
}

type createAnkiCardsData struct {
	Cards []AnkiCard `json:"cards"`
}

// isValidOpenAIModelName checks if the given model name is valid. Needs to be updated when new models are released.
// nolint:gocyclo
func isValidOpenAIModelName(name string) bool {
	switch name {
	case string(GPT432K0613):
	case string(GPT432K0314):
	case string(GPT432K):
	case string(GPT40613):
	case string(GPT40314):
	case string(GPT4o):
	case string(GPT4o20240513):
	case string(GPT4o20240806):
	case string(GPT4oMini):
	case string(GPT4oMini20240718):
	case string(GPT4Turbo):
	case string(GPT4Turbo20240409):
	case string(GPT4Turbo0125):
	case string(GPT4Turbo1106):
	case string(GPT4TurboPreview):
	case string(GPT4VisionPreview):
	case string(GPT4):
	case string(GPT3Dot5Turbo0125):
	case string(GPT3Dot5Turbo1106):
	case string(GPT3Dot5Turbo0613):
	case string(GPT3Dot5Turbo0301):
	case string(GPT3Dot5Turbo16K):
	case string(GPT3Dot5Turbo16K0613):
	case string(GPT3Dot5Turbo):
	case string(GPT3Dot5TurboInstruct):
	case string(GPT3Davinci002):
	case string(GPT3Curie):
	case string(GPT3Curie002):
	case string(GPT3Babbage002):
	case string(TTSModel1):
	case string(TTSModel1HD):
	case string(TTSModelCanary):
	default:
		return false
	}
	return true
}

type OpenAIModelName string

const (
	GPT432K0613           OpenAIModelName = "gpt-4-32k-0613"
	GPT432K0314           OpenAIModelName = "gpt-4-32k-0314"
	GPT432K               OpenAIModelName = "gpt-4-32k"
	GPT40613              OpenAIModelName = "gpt-4-0613"
	GPT40314              OpenAIModelName = "gpt-4-0314"
	GPT4o                 OpenAIModelName = "gpt-4o"
	GPT4o20240513         OpenAIModelName = "gpt-4o-2024-05-13"
	GPT4o20240806         OpenAIModelName = "gpt-4o-2024-08-06"
	GPT4oMini             OpenAIModelName = "gpt-4o-mini"
	GPT4oMini20240718     OpenAIModelName = "gpt-4o-mini-2024-07-18"
	GPT4Turbo             OpenAIModelName = "gpt-4-turbo"
	GPT4Turbo20240409     OpenAIModelName = "gpt-4-turbo-2024-04-09"
	GPT4Turbo0125         OpenAIModelName = "gpt-4-0125-preview"
	GPT4Turbo1106         OpenAIModelName = "gpt-4-1106-preview"
	GPT4TurboPreview      OpenAIModelName = "gpt-4-turbo-preview"
	GPT4VisionPreview     OpenAIModelName = "gpt-4-vision-preview"
	GPT4                  OpenAIModelName = "gpt-4"
	GPT3Dot5Turbo0125     OpenAIModelName = "gpt-3.5-turbo-0125"
	GPT3Dot5Turbo1106     OpenAIModelName = "gpt-3.5-turbo-1106"
	GPT3Dot5Turbo0613     OpenAIModelName = "gpt-3.5-turbo-0613"
	GPT3Dot5Turbo0301     OpenAIModelName = "gpt-3.5-turbo-0301"
	GPT3Dot5Turbo16K      OpenAIModelName = "gpt-3.5-turbo-16k"
	GPT3Dot5Turbo16K0613  OpenAIModelName = "gpt-3.5-turbo-16k-0613"
	GPT3Dot5Turbo         OpenAIModelName = "gpt-3.5-turbo"
	GPT3Dot5TurboInstruct OpenAIModelName = "gpt-3.5-turbo-instruct"
	GPT3Davinci002        OpenAIModelName = "davinci-002"
	GPT3Curie             OpenAIModelName = "curie"
	GPT3Curie002          OpenAIModelName = "curie-002"
	GPT3Babbage002        OpenAIModelName = "babbage-002"
	TTSModel1             OpenAIModelName = "tts-1"
	TTSModel1HD           OpenAIModelName = "tts-1-hd"
	TTSModelCanary        OpenAIModelName = "canary-tts"
)

func (m OpenAIModelName) String() string {
	return string(m)
}

package ai

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var (
	ErrInvalidOpenAIModel = errors.New("invalid openai model")
)

type OpenAIAPIProvider struct {
	model *OpenAIModel
}

func (s *OpenAIAPIProvider) Action() CardCreator {
	return s.model
}

func (s *OpenAIAPIProvider) ModelName() Modeler {
	return s.model.modelType
}

type OpenAIModel struct {
	apiKey    string
	client    *openai.Client
	modelType OpenAIModelName
}

func NewOpenAIModel(apiKey string, modelType ...OpenAIModelName) *OpenAIModel {
	mt := GPT4o20240806
	if len(modelType) > 0 {
		mt = modelType[0]
	}

	client := openai.NewClient(apiKey)

	return &OpenAIModel{
		apiKey:    apiKey,
		modelType: mt,
		client:    client,
	}
}

func (m *OpenAIModel) ChooseDeck(deckNames []string, text string) (string, error) {
	ctx := context.Background()

	deckNameChoices := strings.Join(deckNames, ", ")

	resp, err := m.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: m.modelType.String(),
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
			return "", errors.New("missing key: " + key)
		}
	}

	log.Printf("Result: %v\n", result)
	return deckNames[0], nil
}

func (m *OpenAIModel) CreateAnkiCards(deckName string, text string) ([]AnkiCard, error) {
	return []AnkiCard{{Front: text, Back: text}}, nil
}

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
)

func (m OpenAIModelName) String() string {
	return string(m)
}

package ai

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type ImageGen interface {
	// Generate generates speech from text and returns the audio as an MP3 file.
	Generate(ctx context.Context, text string) ([]byte, error)
}

// ImageGenService is a service for generating text-to-speech audio.
type ImageGenService struct {
	OpenAIClient
}

// NewImageGenService creates a new ImageGen service with the given OpenAI API key.
func NewImageGenService(openAIApiKey string) ImageGen {
	return &ImageGenService{
		*NewOpenAIClient(openAIApiKey, openai.CreateImageModelDallE3),
	}
}

func wordToPrompt(word string) string {
	return fmt.Sprintf("Please create an illustration for the word \"%s\" to help visually represent its meaning for my Anki card.", word)
}

func (svc *ImageGenService) Generate(ctx context.Context, word string) ([]byte, error) {
	raw, err := svc.client.CreateImage(
		ctx,
		openai.ImageRequest{
			Model:          openai.CreateImageModelDallE3,
			Prompt:         wordToPrompt(word),
			N:              1,
			Size:           "1024x1024",
			ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create image: %w", err)
	}

	var bytes []byte
	for _, datum := range raw.Data {
		if datum.B64JSON != "" {
			// Decode base64 string to bytes
			decodedBytes, err := base64.StdEncoding.DecodeString(datum.B64JSON)
			if err != nil {
				return nil, fmt.Errorf("decode base64: %w", err)
			}
			// Append the decoded bytes
			bytes = append(bytes, decodedBytes...)
		}
	}
	return bytes, nil
}

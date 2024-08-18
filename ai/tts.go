package ai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type TTS interface {
	GenerateMP3(text string) ([]byte, error)
}

type TTSService struct {
	apiKey string
	client *openai.Client
}

func NewTTSService(apiKey string) *TTSService {
	client := openai.NewClient(apiKey)

	return &TTSService{
		apiKey: apiKey,
		client: client,
	}
}

func (tts *TTSService) GenerateMP3(text string) (response openai.RawResponse, err error) {
	ctx := context.Background()
	return tts.client.CreateSpeech(
		ctx,
		openai.CreateSpeechRequest{
			Model:          openai.TTSModel1HD,
			Input:          text,
			Voice:          openai.VoiceNova,
			ResponseFormat: openai.SpeechResponseFormatMp3,
			Speed:          0.8,
		},
	)
}

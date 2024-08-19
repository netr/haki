package ai

import (
	"context"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

type TTS interface {
	// GenerateMP3 generates speech from text and returns the audio as an MP3 file.
	GenerateMP3(ctx context.Context, text string) ([]byte, error)
	// GenerateWav generates speech from text and returns the audio as a WAV file.
	GenerateWav(ctx context.Context, text string) ([]byte, error)
	// Generate speech from text. The voice and format can be specified. TODO: make this universal when we add more providers.
	Generate(ctx context.Context, text string, voice openai.SpeechVoice, format openai.SpeechResponseFormat) ([]byte, error)
}

// TTSService is a service for generating text-to-speech audio.
type TTSService struct {
	OpenAIClient
}

// NewTTSService creates a new TTS service with the given OpenAI API key.
func NewTTSService(openAIApiKey string) TTS {
	return &TTSService{
		*NewOpenAIClient(openAIApiKey, TTSModel1),
	}
}

// Generate speech from text. The voice and format can be specified.
// We use the [pause] hack to prevent truncation of audio for some single-word strings.
// https://community.openai.com/t/audio-speech-truncated-audio-for-some-single-word-strings/529924/4
func (tts *TTSService) Generate(ctx context.Context, text string, voice openai.SpeechVoice, format openai.SpeechResponseFormat) ([]byte, error) {
	raw, err := tts.client.CreateSpeech(
		ctx,
		openai.CreateSpeechRequest{
			Model:          openai.SpeechModel(tts.modelType),
			Input:          fmt.Sprintf("\n[pause]\n%s", text),
			Voice:          voice,
			ResponseFormat: format,
			Speed:          1,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create tts request: %w", err)
	}

	var bytes []byte
	bytes, err = io.ReadAll(raw.ReadCloser)
	if err != nil {
		return nil, fmt.Errorf("read tts response: %w", err)
	}
	return bytes, nil
}

// GenerateMP3 generates speech from text and returns the audio as an MP3 file.
func (tts *TTSService) GenerateMP3(ctx context.Context, text string) ([]byte, error) {
	return tts.Generate(ctx, text, openai.VoiceAlloy, openai.SpeechResponseFormatMp3)
}

// GenerateWav generates speech from text and returns the audio as a WAV file.
func (tts *TTSService) GenerateWav(ctx context.Context, text string) ([]byte, error) {
	return tts.Generate(ctx, text, openai.VoiceAlloy, openai.SpeechResponseFormatWav)
}

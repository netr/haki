package ai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type TTS interface {
	GenerateMP3(text string) ([]byte, error)
}

type TTSService struct {
	openAIApiKey string
	client       *openai.Client
}

// NewTTSService creates a new TTS service with the given OpenAI API key.
func NewTTSService(openAIApiKey string) *TTSService {
	client := openai.NewClient(openAIApiKey)

	return &TTSService{
		openAIApiKey: openAIApiKey,
		client:       client,
	}
}

// Generate speech from text. The voice and format can be specified.
// We use the [pause] hack to prevent truncation of audio for some single-word strings.
// https://community.openai.com/t/audio-speech-truncated-audio-for-some-single-word-strings/529924/4
func (tts *TTSService) Generate(text string, voice openai.SpeechVoice, format openai.SpeechResponseFormat) (response openai.RawResponse, err error) {
	ctx := context.Background()
	return tts.client.CreateSpeech(
		ctx,
		openai.CreateSpeechRequest{
			Model:          openai.TTSModel1,
			Input:          fmt.Sprintf("\n[pause]\n%s", text),
			Voice:          voice,
			ResponseFormat: format,
			Speed:          1,
		},
	)
}

// GenerateMP3 generates speech from text and returns the audio as an MP3 file.
func (tts *TTSService) GenerateMP3(text string) (response openai.RawResponse, err error) {
	return tts.Generate(text, openai.VoiceAlloy, openai.SpeechResponseFormatMp3)
}

// GenerateWav generates speech from text and returns the audio as a WAV file.
func (tts *TTSService) GenerateWav(text string) (response openai.RawResponse, err error) {
	return tts.Generate(text, openai.VoiceAlloy, openai.SpeechResponseFormatWav)
}

package ai_test

import (
	"errors"
	"testing"

	"github.com/netr/haki/ai"
)

func Test_NewAPIProvider_WorkingAsExpected(t *testing.T) {
	o, err := ai.NewAPIProvider(ai.OpenAI, "key", ai.GPT3Curie)
	if err != nil || o == nil {
		t.Fatal("openai model provider is nil")
	}

	modelName := o.ModelName()
	if modelName == nil {
		t.Fatal("openai model name is nil")
	}
	if modelName.String() != ai.GPT3Curie.String() {
		t.Fatalf("openai model name is not curie: %s", modelName.String())
	}

	actions := o.Action()
	if actions == nil {
		t.Fatal("openai actions is nil")
	}
}

func Test_NewAIModelProvider_Anthropic_Unimplemented_Fail(t *testing.T) {
	o, err := ai.NewAPIProvider(ai.Anthropic, "key", ai.GPT4Turbo20240409)
	if err == nil || o != nil {
		if errors.Is(err, ai.ErrInvalidAIModelProvider) {
			t.Fatal("anthropic model provider should return ErrUnimplemented")
		}
	}
}

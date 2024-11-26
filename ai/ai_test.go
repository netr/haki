package ai_test

import (
	"errors"
	"testing"

	"github.com/netr/haki/ai"
)

func Test_NewAICardCreator_WorkingAsExpected(t *testing.T) {
	o, err := ai.NewCardCreator(ai.OpenAI, "key", ai.GPT3Curie)
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
}

func Test_NewAICardCreator_Anthropic_Unimplemented_Fail(t *testing.T) {
	o, err := ai.NewCardCreator(ai.Anthropic, "key", ai.GPT4Turbo20240409)
	if err == nil || o != nil {
		if !errors.Is(err, ai.ErrUnimplemented) {
			t.Fatal("anthropic model provider should return ErrUnimplemented")
		}
	}
}

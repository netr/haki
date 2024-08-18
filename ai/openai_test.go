package ai_test

import (
	"testing"

	"github.com/netr/haki/ai"
)

func Test_NewAIModelProvider_OpenAI_DefaultModel(t *testing.T) {
	o, err := ai.NewAPIProvider(ai.OpenAI, "key")
	if err != nil || o == nil {
		t.Fatal("openai model provider is nil")
	}

	modelName := o.ModelName()
	if modelName == nil {
		t.Fatal("openai model name is nil")
	}

	if modelName.String() != ai.GPT4o20240806.String() {
		t.Fatalf("openai model name is not gpt-4-o-2024-08-06: %s", modelName.String())
	}
}

package lib_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sashabaranov/go-openai"

	"github.com/netr/haki/ai"
	"github.com/netr/haki/lib"
)

func TestNewPluginConfigFrom(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "plugin-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatalf("Failed to remove temp directory: %v", err)
		}
	}(tempDir)

	// Create a test config file
	configXML := `<?xml version="1.0" encoding="UTF-8"?>
<plugin>
    <identifier>vocab</identifier>
    <description>Vocabulary cards with audio and images</description>
    <prompt>
        <path>data/prompt.txt</path>
    </prompt>
    <generation>
        <mode>single</mode>
    </generation>
    <deck>
        <filter>Vocabulary::</filter>
    </deck>
    <card_type>
        <model_name>VocabularyWithAudio</model_name>
        <fields>
            <field anki="Question">front</field>
            <field anki="Definition">back</field>
            <field anki="Audio">audio_path</field>
            <field anki="Picture">image_path</field>
        </fields>
    </card_type>
    <services>
        <tts>true</tts>
        <image_gen>true</image_gen>
    </services>
    <output_dir>./data</output_dir>
</plugin>`

	configPath := filepath.Join(tempDir, "config.xml")
	err = os.WriteFile(configPath, []byte(configXML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test successful parsing
	config, err := lib.NewPluginConfigFrom(configPath)
	if err != nil {
		t.Fatalf("Failed to parse valid config: %v", err)
	}

	// Verify parsed values
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Identifier", config.Identifier, "vocab"},
		{"Description", config.Description, "Vocabulary cards with audio and images"},
		{"Text Path", config.Prompt.Path, "data/prompt.txt"},
		{"Generation Mode", config.Generation.Mode, lib.GenerationMode("single")},
		{"Deck Filter", config.Deck.Filter, "Vocabulary::"},
		{"Model Name", config.CardType.ModelName, "VocabularyWithAudio"},
		{"TTS Enabled", config.Services.TTS, true},
		{"ImageGen Enabled", config.Services.ImageGen, true},
		{"Output Directory", config.OutputDir, "./data"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("got %v, want %v", tt.got, tt.expected)
			}
		})
	}

	// Verify field mappings
	fieldMapping := config.GetFieldMapping()
	expectedFields := map[string]string{
		"Question":   "front",
		"Definition": "back",
		"Audio":      "audio_path",
		"Picture":    "image_path",
	}

	for anki, plugin := range expectedFields {
		if fieldMapping[anki] != plugin {
			t.Errorf("Field mapping for %s: got %s, want %s", anki, fieldMapping[anki], plugin)
		}
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectError bool
	}{
		{
			name: "Missing Identifier",
			config: `<?xml version="1.0" encoding="UTF-8"?>
<plugin>
    <prompt><path>data/prompt.txt</path></prompt>
    <generation><mode>single</mode></generation>
    <card_type>
        <model_name>Basic</model_name>
        <fields><field anki="Front">front</field></fields>
    </card_type>
    <output_dir>./data</output_dir>
</plugin>`,
			expectError: true,
		},
		{
			name: "Invalid Generation Mode",
			config: `<?xml version="1.0" encoding="UTF-8"?>
<plugin>
    <identifier>test</identifier>
    <prompt><path>data/prompt.txt</path></prompt>
    <generation><mode>invalid</mode></generation>
    <card_type>
        <model_name>Basic</model_name>
        <fields><field anki="Front">front</field></fields>
    </card_type>
    <output_dir>./data</output_dir>
</plugin>`,
			expectError: true,
		},
		{
			name: "Missing Fields",
			config: `<?xml version="1.0" encoding="UTF-8"?>
<plugin>
    <identifier>test</identifier>
    <prompt><path>data/prompt.txt</path></prompt>
    <generation><mode>single</mode></generation>
    <card_type>
        <model_name>Basic</model_name>
    </card_type>
    <output_dir>./data</output_dir>
</plugin>`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile := filepath.Join(t.TempDir(), "config.xml")
			err := os.WriteFile(tempFile, []byte(tt.config), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			_, err = lib.NewPluginConfigFrom(tempFile)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Mock AI service for testing
type mockAnkiController struct {
	chooseDeckFn        func(ctx context.Context, deckNames []string, text string) (string, error)
	generateAnkiCardsFn func(ctx context.Context, deckName string, text string, prompt string) ([]ai.AnkiCard, error)
}

func (m *mockAnkiController) ChooseDeck(ctx context.Context, deckNames []string, text string) (string, error) {
	return m.chooseDeckFn(ctx, deckNames, text)
}

func (m *mockAnkiController) GenerateAnkiCards(ctx context.Context, deckName string, text string, prompt string) ([]ai.AnkiCard, error) {
	return m.generateAnkiCardsFn(ctx, deckName, text, prompt)
}

func (m *mockAnkiController) ModelName() ai.ModelNamer {
	return ai.GPT4Turbo
}

// Mock TTS service
type mockTTSService struct{}

func (m *mockTTSService) Type() string { return string(lib.ServiceTypeTTS) }
func (m *mockTTSService) GenerateMP3(ctx context.Context, text string) ([]byte, error) {
	return nil, nil
}
func (m *mockTTSService) GenerateWav(ctx context.Context, text string) ([]byte, error) {
	return nil, nil
}
func (m *mockTTSService) Generate(ctx context.Context, text string, voice openai.SpeechVoice, format openai.SpeechResponseFormat) ([]byte, error) {
	return nil, nil
}

// Mock ImageGen service
type mockImageGenService struct{}

func (m *mockImageGenService) Type() string { return string(lib.ServiceTypeImageGen) }
func (m *mockImageGenService) Generate(ctx context.Context, text string) ([]byte, error) {
	return nil, nil
}

func TestNewDerivedPlugin(t *testing.T) {
	// Create a test configuration
	config := &lib.PluginConfig{
		Identifier: "test",
		Services: lib.ServicesConfig{
			TTS:      true,
			ImageGen: true,
		},
		Generation: lib.GenerationConfig{
			Mode: lib.SingleCard,
		},
		CardType: lib.CardTypeConfig{
			ModelName: "Basic",
			Fields: []lib.FieldConfig{
				{AnkiField: "Front", PluginField: "front"},
				{AnkiField: "Back", PluginField: "back"},
			},
		},
	}

	tests := []struct {
		name          string
		services      []lib.PluginService
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid configuration with all required services",
			services: []lib.PluginService{
				&mockTTSService{},
				&mockImageGenService{},
			},
			expectError: false,
		},
		{
			name:          "Missing required TTS service",
			services:      []lib.PluginService{&mockImageGenService{}},
			expectError:   true,
			errorContains: "required service tts not provided",
		},
		{
			name:          "Missing required ImageGen service",
			services:      []lib.PluginService{&mockTTSService{}},
			expectError:   true,
			errorContains: "required service image_gen not provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAI := &mockAnkiController{}
			plugin, err := lib.NewDerivedPlugin(config, mockAI, tt.services...)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if plugin == nil {
					t.Error("expected non-nil plugin")
				}
			}
		})
	}
}

func TestDerivedPluginGenerateCards(t *testing.T) {
	config := &lib.PluginConfig{
		Generation: lib.GenerationConfig{
			Mode: lib.SingleCard,
		},
		Prompt: lib.PromptConfig{
			Path: "data/prompt.txt",
		},
		CardType: lib.CardTypeConfig{
			ModelName: "Basic",
			Fields: []lib.FieldConfig{
				{AnkiField: "Front", PluginField: "front"},
				{AnkiField: "Back", PluginField: "back"},
			},
		},
	}

	tests := []struct {
		name          string
		cards         []ai.AnkiCard
		expectError   bool
		errorContains string
	}{
		{
			name: "Single card generation success",
			cards: []ai.AnkiCard{
				{Front: "Test Front", Back: "Test Back"},
			},
			expectError: false,
		},
		{
			name: "Multiple cards when single mode",
			cards: []ai.AnkiCard{
				{Front: "Card 1 Front", Back: "Card 1 Back"},
				{Front: "Card 2 Front", Back: "Card 2 Back"},
			},
			expectError:   true,
			errorContains: "plugin configured for single card but received multiple cards",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAI := &mockAnkiController{
				generateAnkiCardsFn: func(ctx context.Context, deckName string, text string, prompt string) ([]ai.AnkiCard, error) {
					return tt.cards, nil
				},
			}

			plugin, err := lib.NewDerivedPlugin(config, mockAI)
			if err != nil {
				t.Fatalf("Failed to create plugin: %v", err)
			}

			cards, err := plugin.GenerateAnkiCards(context.Background(), "test query")

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(cards) != len(tt.cards) {
					t.Errorf("expected %d cards, got %d", len(tt.cards), len(cards))
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) > len(substr) && s[len(s)-len(substr):] == substr
}

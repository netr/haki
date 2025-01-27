package lib

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/netr/haki/ai"
	"github.com/netr/haki/anki"
)

// PromptConfig defines the prompt configuration
type PromptConfig struct {
	Path string `xml:"path"`
}

// GenerationConfig defines the generation mode configuration
type GenerationConfig struct {
	Mode GenerationMode `xml:"mode"`
}

// GenerationMode represents the card generation mode
type GenerationMode string

const (
	SingleCard    GenerationMode = "single"
	MultipleCards GenerationMode = "multiple"
)

// DeckConfig defines the deck configuration
type DeckConfig struct {
	Filter string `xml:"filter"`
}

// CardTypeConfig defines the card type configuration
type CardTypeConfig struct {
	ModelName string        `xml:"model_name"`
	Fields    []FieldConfig `xml:"fields>field"`
}

// FieldConfig represents a field mapping configuration
type FieldConfig struct {
	AnkiField   string `xml:"anki,attr"`
	PluginField string `xml:",chardata"`
}

// ServicesConfig defines which services are enabled
type ServicesConfig struct {
	TTS      bool `xml:"tts"`
	ImageGen bool `xml:"image_gen"`
}

// PluginConfig represents the complete configuration for a plugin
type PluginConfig struct {
	Identifier  string           `xml:"identifier"`
	Description string           `xml:"description"`
	Prompt      PromptConfig     `xml:"prompt"`
	Generation  GenerationConfig `xml:"generation"`
	Deck        DeckConfig       `xml:"deck"`
	CardType    CardTypeConfig   `xml:"card_type"`
	Services    ServicesConfig   `xml:"services"`
	OutputDir   string           `xml:"output_dir"`
}

// NewPluginConfigFrom loads and parses a plugin configuration from the given file path
func NewPluginConfigFrom(outputDir, path string) (*PluginConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading plugin config file: %w", err)
	}

	// Create a new config instance
	config := &PluginConfig{}

	// Parse the XML
	if err := xml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("parsing plugin config: %w", err)
	}

	// Validate the configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("validating plugin config: %w", err)
	}

	config.OutputDir = outputDir
	return config, nil
}

// validate performs validation of the plugin configuration
func (c *PluginConfig) validate() error {
	if c.Identifier == "" {
		return ErrPluginIdentifierRequired
	}

	if c.Prompt.Path == "" {
		return ErrPromptPathRequired
	}

	if c.Generation.Mode == "" {
		return ErrGenerationModeRequired
	}

	if c.Generation.Mode != SingleCard && c.Generation.Mode != MultipleCards {
		return &ErrInvalidGenerationMode{Mode: c.Generation.Mode}
	}

	if c.CardType.ModelName == "" {
		return ErrCardTypeModelNameRequired
	}

	if len(c.CardType.Fields) == 0 {
		return ErrFieldMappingRequired
	}

	if c.OutputDir == "" {
		return ErrOutputDirectoryRequired
	}

	// Keep the MkdirAll error as is since it's wrapping a system error
	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	return nil
}

// GetPromptContentFrom loads the prompt content from the configured path
func (c *PluginConfig) GetPromptContentFrom(outputDir string) (string, error) {
	promptPath := c.Prompt.Path
	if !filepath.IsAbs(promptPath) {
		if !strings.HasSuffix(outputDir, "/") {
			outputDir += "/"
		}
		// If the path is relative, resolve it relative to the config file location
		configDir := filepath.Dir(outputDir)
		promptPath = filepath.Join(configDir, promptPath)
	}

	content, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("reading prompt file: %w", err)
	}

	return string(content), nil
}

// GetFieldMapping returns a map of Anki field names to plugin field names
func (c *PluginConfig) GetFieldMapping() map[string]string {
	mapping := make(map[string]string)
	for _, field := range c.CardType.Fields {
		mapping[field.AnkiField] = field.PluginField
	}
	return mapping
}

var (
	ErrQueryRequired = fmt.Errorf("query is required")
)

// BasePlugin is a base implementation of the AnkiCardGeneratorPlugin interface
type BasePlugin struct {
	ankiClient anki.AnkiClienter
	ankiAI     ai.AnkiController
	deckName   string
	aiServices BaseAIServices
}

type BaseAIServices struct {
	tts      ai.TTS
	imageGen ai.ImageGen
}

// NewBasePlugin creates a new instance of the BasePlugin
func NewBasePlugin(c ai.AnkiController) *BasePlugin {
	return &BasePlugin{
		ankiClient: anki.NewClient(GetEnv("ANKI_CONNECT_URL", "http://localhost:8765")),
		ankiAI:     c,
	}
}

func (t *BasePlugin) getFilteredDeckNames(filter ...string) ([]string, error) {
	deckNames, err := t.getBaseDeckNames()
	if err != nil {
		return nil, fmt.Errorf("get filtered deck names: %w", err)
	}

	decks, err := t.filterDecksByName(deckNames, filter...)
	if err != nil {
		return nil, fmt.Errorf("get filtered deck names: %w", err)
	}
	return decks, nil
}

// getBaseDeckNames - this is an very big assumption that people who use Anki have sub folders.
// This returns all decks that don't have `::` in the name.
func (t *BasePlugin) getBaseDeckNames() ([]string, error) {
	deckNames, err := t.ankiClient.DeckNames().GetNames()
	if err != nil {
		return nil, fmt.Errorf("get base deck names: %w", err)
	}
	return anki.FilterDecksByHierarchy(deckNames), nil
}

func (t *BasePlugin) filterDecksByName(deckNames []string, filter ...string) ([]string, error) {
	if len(filter) == 0 {
		return deckNames, nil
	}

	// only use deck names that have [filter] in them
	var decks []string
	for _, d := range deckNames {
		if strings.Contains(d, filter[0]) {
			decks = append(decks, d)
		}
	}
	return decks, nil
}

func (t *BasePlugin) generateAnkiCards(ctx context.Context, query string, prompt string) ([]ai.AnkiCard, error) {
	cards, err := t.ankiAI.GenerateAnkiCards(
		ctx,
		t.deckName,
		query,
		prompt,
	)
	if err != nil {
		return nil, fmt.Errorf("generate anki cards: %w", err)
	}
	return cards, nil
}

func (t *BasePlugin) chooseDeck(ctx context.Context, query string, decks []string, createIfNotExists bool) (string, error) {
	deckName, err := t.ankiAI.ChooseDeck(ctx, decks, fmt.Sprintf("Which deck should I use for the topic: %s", query))
	if err != nil {
		return "", err
	}

	if createIfNotExists {
		if !slices.Contains(decks, deckName) {
			if err := t.ankiClient.DeckNames().Create(deckName); err != nil {
				return "", err
			}
		}
	}

	t.deckName = deckName
	return deckName, nil
}

func (t *BasePlugin) generateTTS(ctx context.Context, query, outputDir string) (string, error) {
	mp3Bytes, err := t.aiServices.tts.GenerateMP3(ctx, query)
	if err != nil {
		return "", fmt.Errorf("generate mp3: %w", err)
	}
	path, err := makeTTSFilePath(outputDir, query)
	if err != nil {
		return "", fmt.Errorf("make file path: %w", err)
	}
	err = SaveFile(path, mp3Bytes)
	if err != nil {
		return "", fmt.Errorf("save file: %w", err)
	}
	return path, nil
}

func makeTTSFilePath(outputDir, word string) (string, error) {
	return filepath.Abs(fmt.Sprintf("%s/data/%s.mp3", outputDir, word))
}

func (t *BasePlugin) generateImage(ctx context.Context, query, outputDir string) (string, error) {
	imgBytes, err := t.aiServices.imageGen.Generate(ctx, query)
	if err != nil {
		return "", fmt.Errorf("generate image: %w", err)
	}
	path, err := makeImageFilePath(outputDir, query)
	if err != nil {
		return "", fmt.Errorf("make file path: %w", err)
	}
	err = SaveFile(path, imgBytes)
	if err != nil {
		return "", fmt.Errorf("save file: %w", err)
	}
	return path, nil
}
func makeImageFilePath(outputDir, word string) (string, error) {
	return filepath.Abs(fmt.Sprintf("%s/data/%s.webp", outputDir, word))
}

func PrintCards(ac []ai.AnkiCard, padding bool) {
	if padding {
		fmt.Println("")
	}
	for _, c := range ac {
		fmt.Print(Colors.BeautifyCard(c))
	}
	if padding {
		fmt.Println("")
	}
}

// DerivedPlugin combines the BasePlugin with configuration settings
type DerivedPlugin struct {
	*BasePlugin
	config *PluginConfig
}

// NewDerivedPlugin creates a new DerivedPlugin instance
func NewDerivedPlugin(config *PluginConfig, cardCreator ai.AnkiController, services ...PluginService) (*DerivedPlugin, error) {
	if err := validateServices(config, services); err != nil {
		return nil, fmt.Errorf("validating services: %w", err)
	}

	base := NewBasePlugin(cardCreator)
	derived := &DerivedPlugin{
		BasePlugin: base,
		config:     config,
	}

	// Attach services based on configuration
	if err := derived.attachServices(services); err != nil {
		return nil, fmt.Errorf("attaching services: %w", err)
	}

	return derived, nil
}

// PluginService represents an optional service that can be attached to a plugin
type PluginService interface {
	Type() string
}

type ServiceType string

const (
	ServiceTypeTTS      ServiceType = "tts"
	ServiceTypeImageGen ServiceType = "image_gen"
)

func ServiceTypeFrom(s string) ServiceType {
	if s == string(ServiceTypeTTS) {
		return ServiceTypeTTS
	}
	if s == string(ServiceTypeImageGen) {
		return ServiceTypeImageGen
	}
	return ""
}

func validateServices(config *PluginConfig, services []PluginService) error {
	// Create a map of required services from config
	required := map[ServiceType]bool{
		ServiceTypeTTS:      config.Services.TTS,
		ServiceTypeImageGen: config.Services.ImageGen,
	}

	// Create a map of provided services
	provided := make(map[ServiceType]bool)
	for _, service := range services {
		if service == nil {
			continue
		}
		s := ServiceTypeFrom(service.Type())
		if s == "" {
			return &ErrUnknownServiceType{ServiceType: service.Type()}
		}
		provided[s] = true
	}

	// Check if all required services are provided
	for serviceType, isRequired := range required {
		if isRequired && !provided[serviceType] {
			return &ErrRequiredServiceNotProvided{ServiceType: serviceType}
		}
	}

	return nil
}

func (d *DerivedPlugin) attachServices(services []PluginService) error {
	for _, service := range services {
		if service == nil {
			continue
		}
		switch ServiceTypeFrom(service.Type()) {
		case ServiceTypeTTS:
			if tts, ok := service.(ai.TTS); ok {
				d.aiServices.tts = tts
			} else {
				return ErrInvalidTTSService
			}
		case ServiceTypeImageGen:
			if imageGen, ok := service.(ai.ImageGen); ok {
				d.aiServices.imageGen = imageGen
			} else {
				return ErrInvalidImageGenService
			}
		default:
			return &ErrUnknownServiceType{ServiceType: service.Type()}
		}
	}
	return nil
}

// ChooseDeck implements AnkiCardGeneratorPlugin interface
func (d *DerivedPlugin) ChooseDeck(ctx context.Context, query string) (string, error) {
	if query == "" {
		return "", ErrQueryRequired
	}

	decks, err := d.getFilteredDeckNames(d.config.Deck.Filter)
	if err != nil {
		return "", err
	}

	deckName, err := d.chooseDeck(ctx, query, decks, true)
	if err != nil {
		return "", fmt.Errorf("choose filtered deck: %w", err)
	}

	return deckName, nil
}

// GenerateAnkiCards implements AnkiCardGeneratorPlugin interface
func (d *DerivedPlugin) GenerateAnkiCards(ctx context.Context, prompt string, query string) ([]ai.AnkiCard, error) {
	cards, err := d.generateAnkiCards(ctx, query, prompt)
	if err != nil {
		return nil, err
	}

	if err := d.validateGeneratedCards(cards); err != nil {
		return nil, fmt.Errorf("validate generated cards: %w", err)
	}

	return cards, nil
}

func (d *DerivedPlugin) validateGeneratedCards(cards []ai.AnkiCard) error {
	if d.config.Generation.Mode == SingleCard && len(cards) > 1 {
		return ErrSingleCardModeViolation
	}
	return nil
}

// StoreAnkiCards implements AnkiCardGeneratorPlugin interface
func (d *DerivedPlugin) StoreAnkiCards(deckName string, query string, cards []ai.AnkiCard, allowDuplicates bool) error {
	fieldMapping := d.config.GetFieldMapping()

	for _, card := range cards {
		data := make(map[string]interface{})

		// Map card fields according to configuration
		for ankiField, pluginField := range fieldMapping {
			switch pluginField {
			case "front":
				data[ankiField] = card.Front
			case "back":
				data[ankiField] = formatBack(card.Back)
			}
		}

		note := anki.NewNoteBuilder(deckName, d.config.CardType.ModelName, data, allowDuplicates)

		// Add any additional services
		for ankiField, pluginField := range fieldMapping {
			switch pluginField {
			case "audio_front":
				if d.aiServices.tts != nil {
					startTime := time.Now()
					ttsPath, err := d.generateTTS(context.Background(), query, d.config.OutputDir)
					if err != nil {
						slog.Error("vocab: gen tts", slog.String("error", err.Error()))
						continue
					}
					note.SetField(ankiField, createAudioTag(makeTTSFileName(query))).
						WithAudio(ttsPath, makeTTSFileName(query), "Front")

					slog.Info("audio generated",
						slog.String("deck", deckName),
						slog.String("model", d.config.CardType.ModelName),
						slog.String("duration", time.Since(startTime).String()),
					)
				}
			case "image_front":
				startTime := time.Now()
				imgPath, err := d.generateImage(context.Background(), query, d.config.OutputDir)
				if err != nil {
					slog.Error("vocab: gen image", slog.String("error", err.Error()))
					continue
				}
				note.SetField(ankiField, createImageTag(makeImageFileName(query))).
					WithPicture("", imgPath, makeImageFileName(query), "Front")

				slog.Info("image generated",
					slog.String("deck", deckName),
					slog.String("model", d.config.CardType.ModelName),
					slog.String("duration", time.Since(startTime).String()),
				)
			}
		}

		id, err := d.ankiClient.Notes().Add(note.Build())
		if err != nil {
			return err
		}

		slog.Info("note added",
			slog.String("deck", deckName),
			slog.String("model", d.config.CardType.ModelName),
			slog.String("id", fmt.Sprintf("%.f", id)),
		)
	}

	return nil
}

func createImageTag(fileName string) string {
	return fmt.Sprintf("<img src=\"%s\">", fileName)
}

func createAudioTag(fileName string) string {
	return fmt.Sprintf("[sound:%s]", fileName)
}

func makeImageFileName(name string) string {
	if strings.HasSuffix(name, ".webp") {
		return name
	}
	return fmt.Sprintf("%s.webp", name)
}

func makeTTSFileName(name string) string {
	if strings.HasSuffix(name, ".mp3") {
		return name
	}
	return fmt.Sprintf("%s.mp3", name)
}

func replaceBold(text string) string {
	re := regexp.MustCompile(`\*\*(.*?)\*\*`)
	return re.ReplaceAllString(text, `<b>$1</b>`)
}

func formatBack(data string) string {
	if !strings.Contains(strings.ToLower(data), "<br>") {
		data = strings.ReplaceAll(data, "\n", "<br>\n")
	}
	data = strings.ReplaceAll(data, "\\<", "<")
	data = replaceBold(data)
	return data
}

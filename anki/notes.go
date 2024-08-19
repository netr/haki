package anki

import "fmt"

type NoteService struct {
	client *Client
}

func NewNoteService(client *Client) *NoteService {
	return &NoteService{client}
}

func (svc *NoteService) Add(note Note) (float64, error) {
	var id float64
	if err := svc.client.sendAndUnmarshal("addNote", NoteParams{Note: note}, &id); err != nil {
		return 0, fmt.Errorf("addNote: %w", err)
	}
	return id, nil
}

// NoteParams contains the parameters for adding a new note
type NoteParams struct {
	Note Note `json:"note"`
}

// Note represents an Anki note
type Note struct {
	DeckName  string                 `json:"deckName"`
	ModelName string                 `json:"modelName"`
	Fields    map[string]interface{} `json:"fields"`
	Options   NoteOptions            `json:"2options"`
	Tags      []string               `json:"tags"`
	Audio     []NoteMedia            `json:"audio"`
	Video     []NoteMedia            `json:"video"`
	Picture   []NoteMedia            `json:"picture"`
}

// NoteOptions represents the options for adding a note
type NoteOptions struct {
	AllowDuplicate        bool                  `json:"allowDuplicate"`
	DuplicateScope        string                `json:"duplicateScope"`
	DuplicateScopeOptions DuplicateScopeOptions `json:"duplicateScopeOptions"`
}

// DuplicateScopeOptions represents the options for duplicate scope
type DuplicateScopeOptions struct {
	DeckName       string `json:"deckName"`
	CheckChildren  bool   `json:"checkChildren"`
	CheckAllModels bool   `json:"checkAllModels"`
}

// NoteMedia represents media (audio, video, picture) attached to a note
type NoteMedia struct {
	Path     string   `json:"path"`
	URL      string   `json:"url"`
	Filename string   `json:"filename"`
	SkipHash string   `json:"skipHash"`
	Fields   []string `json:"fields"`
}

// NoteBuilder facilitates the creation of Note structs with sensible defaults
type NoteBuilder struct {
	note Note
}

// NewNoteBuilder creates a new NoteBuilder with minimum required fields and sensible defaults
func NewNoteBuilder(deckName, modelName string, fields map[string]interface{}) *NoteBuilder {
	return &NoteBuilder{
		note: Note{
			DeckName:  deckName,
			ModelName: modelName,
			Fields:    fields,
			Options: NoteOptions{
				AllowDuplicate: false,
				DuplicateScope: "deck",
				DuplicateScopeOptions: DuplicateScopeOptions{
					DeckName:       deckName,
					CheckChildren:  false,
					CheckAllModels: false,
				},
			},
			Tags: []string{},
		},
	}
}

// WithTags adds tags to the note
func (nb *NoteBuilder) WithTags(tags ...string) *NoteBuilder {
	nb.note.Tags = append(nb.note.Tags, tags...)
	return nb
}

// WithAudio adds an audio file to the note
func (nb *NoteBuilder) WithAudio(path, filename string, fields ...string) *NoteBuilder {
	nb.note.Audio = append(nb.note.Audio, NoteMedia{
		Path:     path,
		Filename: filename,
		Fields:   fields,
	})
	return nb
}

// WithVideo adds a video file to the note
func (nb *NoteBuilder) WithVideo(url, filename string, fields ...string) *NoteBuilder {
	nb.note.Video = append(nb.note.Video, NoteMedia{
		URL:      url,
		Filename: filename,
		Fields:   fields,
	})
	return nb
}

// WithPicture adds a picture file to the note
func (nb *NoteBuilder) WithPicture(url, filename string, fields ...string) *NoteBuilder {
	nb.note.Picture = append(nb.note.Picture, NoteMedia{
		URL:      url,
		Filename: filename,
		Fields:   fields,
	})
	return nb
}

// AllowDuplicate sets whether to allow duplicate notes
func (nb *NoteBuilder) AllowDuplicate(allow bool) *NoteBuilder {
	nb.note.Options.AllowDuplicate = allow
	return nb
}

// SetDuplicateScope sets the scope for checking duplicates
func (nb *NoteBuilder) SetDuplicateScope(scope string) *NoteBuilder {
	nb.note.Options.DuplicateScope = scope
	return nb
}

// Build finalizes and returns the Note struct
func (nb *NoteBuilder) Build() Note {
	return nb.note
}

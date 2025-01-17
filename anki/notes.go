package anki

import "fmt"

type NoteService struct {
	client *Client
}

func NewNoteService(client *Client) *NoteService {
	return &NoteService{client}
}

// NoteParams contains the parameters for adding a new note
type AddNoteParams struct {
	Note Note `json:"note"`
}

func (svc *NoteService) Add(note Note) (float64, error) {
	var id float64
	if err := svc.client.sendAndUnmarshal("addNote", AddNoteParams{Note: note}, &id); err != nil {
		return 0, fmt.Errorf("add note: %w", err)
	}
	return id, nil
}

type FindNoteParams struct {
	Query string `json:"query"`
}

func (svc *NoteService) FindNotes(query string) ([]int64, error) {
	var res []int64
	if err := svc.client.sendAndUnmarshal("findNotes", FindNoteParams{Query: query}, &res); err != nil {
		return nil, fmt.Errorf("find notes: %w", err)
	}

	return res, nil
}

type GetNoteInfosParams struct {
	Notes []int64 `json:"notes"`
}

func (svc *NoteService) GetNoteInfos(noteIDs []int64) ([]NoteInfoResult, error) {
	var res []NoteInfoResult
	if err := svc.client.sendAndUnmarshal("notesInfo", GetNoteInfosParams{Notes: noteIDs}, &res); err != nil {
		return nil, fmt.Errorf("get note infos: %w", err)
	}

	return res, nil
}

type DeleteNotesParams struct {
	Notes []int64 `json:"notes"`
}

func (svc *NoteService) DeleteNotes(noteIDs []int64) error {
	var res interface{}
	if err := svc.client.sendAndUnmarshal("deleteNotes", DeleteNotesParams{Notes: noteIDs}, &res); err != nil {
		return fmt.Errorf("delete notes: %w", err)
	}

	return nil
}

type FindNotesResult struct {
	NoteIDs []int64
}

type NoteInfoResult struct {
	NoteID    int64                    `json:"noteId"`
	Profile   string                   `json:"profile"`
	ModelName string                   `json:"modelName"`
	Tags      []string                 `json:"tags"`
	Fields    map[string]NoteInfoField `json:"fields"`
	Mod       int64                    `json:"mod"`
	Cards     []int64                  `json:"cards"`
}

func (r NoteInfoResult) HasField(field string) bool {
	if val, ok := r.Fields[field]; !ok {
		return false
	} else if val.Value == "" {
		return false
	}

	return true
}

type NoteInfoField struct {
	Value string `json:"value"`
	Order int64  `json:"order"`
}

// Note represents an Anki note
type Note struct {
	DeckName  string                 `json:"deckName"`
	ModelName string                 `json:"modelName"`
	Fields    map[string]interface{} `json:"fields"`
	Options   NoteOptions            `json:"options"`
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
// TODO: Path and URL are mutually exclusive, but we can't enforce that in Go. GenerateAnkiCards a better way of handling this and propagate the change to the API
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
func NewNoteBuilder(deckName, modelName string, fields map[string]interface{}, allowDuplicates bool) *NoteBuilder {
	return &NoteBuilder{
		note: Note{
			DeckName:  deckName,
			ModelName: modelName,
			Fields:    fields,
			Options: NoteOptions{
				AllowDuplicate: allowDuplicates,
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
func (nb *NoteBuilder) WithPicture(url, path, filename string, fields ...string) *NoteBuilder {
	if filename == "" {
		return nb
	}
	if len(fields) == 0 {
		fields = []string{"Front"}
	}
	nb.note.Picture = append(nb.note.Picture, NoteMedia{
		URL:      url,
		Path:     path,
		Filename: filename,
		Fields:   fields,
	})
	return nb
}

// SetField appends a field to the note
func (nb *NoteBuilder) SetField(name string, value interface{}) *NoteBuilder {
	nb.note.Fields[name] = value
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

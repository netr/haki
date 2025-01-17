package lib

import (
	"github.com/netr/haki/anki"
)

type VocabUpdaterService struct {
	client      *anki.Client
	noteService *anki.NoteService
}

func NewVocabUpdaterService(client *anki.Client) *VocabUpdaterService {
	return &VocabUpdaterService{
		client:      client,
		noteService: anki.NewNoteService(client),
	}
}

// GetUpdatableNotes returns the noteIDs of the vocabulary cards that need updating
func (s VocabUpdaterService) GetUpdatableNotes(findNotesQuery, expectedModel string, requiredFields []string) ([]anki.NoteInfoResult, error) {
	notes, err := s.noteService.FindNotes(findNotesQuery)
	if err != nil {
		return nil, err
	}

	infos, err := s.noteService.GetNoteInfos(notes)
	if err != nil {
		return nil, err
	}

	var res []anki.NoteInfoResult
	for _, r := range infos {
		if r.ModelName != expectedModel {
			res = append(res, r)
		} else {
			for _, s := range requiredFields {
				if !r.HasField(s) {
					res = append(res, r)
					break
				}
			}
		}
	}

	return res, nil
}

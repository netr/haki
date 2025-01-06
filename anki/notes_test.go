package anki_test

import (
	"reflect"
	"testing"

	"github.com/netr/haki/anki"
)

func TestNewNoteBuilder(t *testing.T) {
	builder := anki.NewNoteBuilder("TestDeck", "BasicModel", map[string]interface{}{"Front": "Front Content", "Back": "Back Content"})
	note := builder.Build()

	if note.DeckName != "TestDeck" {
		t.Errorf("Expected DeckName to be 'TestDeck', got '%s'", note.DeckName)
	}
	if note.ModelName != "BasicModel" {
		t.Errorf("Expected ModelName to be 'BasicModel', got '%s'", note.ModelName)
	}
	if note.Fields["Front"] != "Front Content" {
		t.Errorf("Expected Front field to be 'Front Content', got '%s'", note.Fields["Front"])
	}
	if note.Fields["Back"] != "Back Content" {
		t.Errorf("Expected Back field to be 'Back Content', got '%s'", note.Fields["Back"])
	}
	if note.Options.AllowDuplicate != false {
		t.Errorf("Expected AllowDuplicate to be false by default")
	}
	if note.Options.DuplicateScope != "deck" {
		t.Errorf("Expected DuplicateScope to be 'deck' by default, got '%s'", note.Options.DuplicateScope)
	}
	if len(note.Tags) != 0 {
		t.Errorf("Expected Tags to be empty by default")
	}
}

func TestWithTags(t *testing.T) {
	builder := anki.NewNoteBuilder("TestDeck", "BasicModel", map[string]interface{}{"Front": "Front Content", "Back": "Back Content"}).
		WithTags("tag1", "tag2")
	note := builder.Build()

	expectedTags := []string{"tag1", "tag2"}
	if !reflect.DeepEqual(note.Tags, expectedTags) {
		t.Errorf("Expected Tags to be %v, got %v", expectedTags, note.Tags)
	}
}

func TestWithAudio(t *testing.T) {
	builder := anki.NewNoteBuilder("TestDeck", "BasicModel", map[string]interface{}{"Front": "Front Content", "Back": "Back Content"}).
		WithAudio("http://example.com/audio.mp3", "audio.mp3", "Back")
	note := builder.Build()

	if len(note.Audio) != 1 {
		t.Fatalf("Expected 1 audio file, got %d", len(note.Audio))
	}
	audio := note.Audio[0]
	if audio.Path != "http://example.com/audio.mp3" {
		t.Errorf("Expected audio Path to be 'http://example.com/audio.mp3', got '%s'", audio.URL)
	}
	if audio.Filename != "audio.mp3" {
		t.Errorf("Expected audio filename to be 'audio.mp3', got '%s'", audio.Filename)
	}
	if !reflect.DeepEqual(audio.Fields, []string{"Back"}) {
		t.Errorf("Expected audio fields to be ['Back'], got %v", audio.Fields)
	}
}

func TestWithVideo(t *testing.T) {
	builder := anki.NewNoteBuilder("TestDeck", "BasicModel", map[string]interface{}{"Front": "Front Content", "Back": "Back Content"}).
		WithVideo("http://example.com/video.mp4", "video.mp4", "Front")
	note := builder.Build()

	if len(note.Video) != 1 {
		t.Fatalf("Expected 1 video file, got %d", len(note.Video))
	}
	video := note.Video[0]
	if video.URL != "http://example.com/video.mp4" {
		t.Errorf("Expected video URL to be 'http://example.com/video.mp4', got '%s'", video.URL)
	}
	if video.Filename != "video.mp4" {
		t.Errorf("Expected video filename to be 'video.mp4', got '%s'", video.Filename)
	}
	if !reflect.DeepEqual(video.Fields, []string{"Front"}) {
		t.Errorf("Expected video fields to be ['Front'], got %v", video.Fields)
	}
}

func TestWithPicture(t *testing.T) {
	builder := anki.NewNoteBuilder("TestDeck", "BasicModel", map[string]interface{}{"Front": "Front Content", "Back": "Back Content"}).
		WithPicture("http://example.com/image.jpg", "image.jpg", "image.jpg", "Front", "Back")
	note := builder.Build()

	if len(note.Picture) != 1 {
		t.Fatalf("Expected 1 picture file, got %d", len(note.Picture))
	}
	picture := note.Picture[0]
	if picture.URL != "http://example.com/image.jpg" {
		t.Errorf("Expected picture URL to be 'http://example.com/image.jpg', got '%s'", picture.URL)
	}
	if picture.Filename != "image.jpg" {
		t.Errorf("Expected picture filename to be 'image.jpg', got '%s'", picture.Filename)
	}
	if !reflect.DeepEqual(picture.Fields, []string{"Front", "Back"}) {
		t.Errorf("Expected picture fields to be ['Front', 'Back'], got %v", picture.Fields)
	}
}

func TestAllowDuplicate(t *testing.T) {
	builder := anki.NewNoteBuilder("TestDeck", "BasicModel", map[string]interface{}{"Front": "Front Content", "Back": "Back Content"}).
		AllowDuplicate(true)
	note := builder.Build()

	if !note.Options.AllowDuplicate {
		t.Errorf("Expected AllowDuplicate to be true, got false")
	}
}

func TestSetDuplicateScope(t *testing.T) {
	builder := anki.NewNoteBuilder("TestDeck", "BasicModel", map[string]interface{}{"Front": "Front Content", "Back": "Back Content"}).
		SetDuplicateScope("collection")
	note := builder.Build()

	if note.Options.DuplicateScope != "collection" {
		t.Errorf("Expected DuplicateScope to be 'collection', got '%s'", note.Options.DuplicateScope)
	}
}

func TestMultipleMediaAdditions(t *testing.T) {
	builder := anki.NewNoteBuilder("TestDeck", "BasicModel", map[string]interface{}{"Front": "Front Content", "Back": "Back Content"}).
		WithAudio("http://example.com/audio1.mp3", "audio1.mp3", "Back").
		WithAudio("http://example.com/audio2.mp3", "audio2.mp3", "Front").
		WithVideo("http://example.com/video1.mp4", "video1.mp4", "Front").
		WithPicture("http://example.com/image1.jpg", "", "image1.jpg", "Back")
	note := builder.Build()

	if len(note.Audio) != 2 {
		t.Errorf("Expected 2 audio files, got %d", len(note.Audio))
	}
	if len(note.Video) != 1 {
		t.Errorf("Expected 1 video file, got %d", len(note.Video))
	}
	if len(note.Picture) != 1 {
		t.Errorf("Expected 1 picture file, got %d", len(note.Picture))
	}
}

// Edge case tests

func TestEmptyFields(t *testing.T) {
	builder := anki.NewNoteBuilder("TestDeck", "BasicModel", map[string]interface{}{"Front": "", "Back": ""})
	note := builder.Build()

	if note.Fields["Front"] != "" || note.Fields["Back"] != "" {
		t.Errorf("Expected empty Front and Back fields")
	}
}

func TestDuplicateScepparams(t *testing.T) {
	builder := anki.NewNoteBuilder("TestDeck", "BasicModel", map[string]interface{}{"Front": "Front Content", "Back": "Back Content"}).
		SetDuplicateScope("invalid_scope")
	note := builder.Build()

	if note.Options.DuplicateScope != "invalid_scope" {
		t.Errorf("Expected DuplicateScope to be 'invalid_scope', got '%s'", note.Options.DuplicateScope)
	}
}

func TestMultipleTagAdditions(t *testing.T) {
	builder := anki.NewNoteBuilder("TestDeck", "BasicModel", map[string]interface{}{"Front": "Front Content", "Back": "Back Content"}).
		WithTags("tag1", "tag2").
		WithTags("tag3")
	note := builder.Build()

	expectedTags := []string{"tag1", "tag2", "tag3"}
	if !reflect.DeepEqual(note.Tags, expectedTags) {
		t.Errorf("Expected Tags to be %v, got %v", expectedTags, note.Tags)
	}
}

func TestMediaWithNoFields(t *testing.T) {
	builder := anki.NewNoteBuilder("TestDeck", "BasicModel", map[string]interface{}{"Front": "Front Content", "Back": "Back Content"}).
		WithAudio("http://example.com/audio.mp3", "audio.mp3")
	note := builder.Build()

	if len(note.Audio) != 1 {
		t.Fatalf("Expected 1 audio file, got %d", len(note.Audio))
	}
	if len(note.Audio[0].Fields) != 0 {
		t.Errorf("Expected no fields for audio, got %v", note.Audio[0].Fields)
	}
}

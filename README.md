# Anki Card Creator

This tool uses the AnkiConnect API and AI APIs to create Anki cards efficiently. The motivation behind this tool is to make the process of creating Anki cards a lot easier and faster. When you're reading a book and want to create Anki cards for the new words you come across, you can use this tool to create the cards in a few seconds and continue reading.

## Installation

```bash
go install github.com/netr/haki@latest
```

## Features / Commands

### Vocabulary Cards
```bash
haki vocab --word "cacophony"
```

- [x] Automatically fetches the definition and example sentence.
- [x] Creates a TTS of the word using OpenAI's tts-1 model.
- [ ] Automatically fetch the pronunciation of the word.



package lib

import (
	"fmt"

	"github.com/netr/haki/ai"
)

type AnsiColors struct {
	Reset  string
	Red    string
	Green  string
	Yellow string
	Blue   string
	Purple string
	Cyan   string
	White  string
}

var Colors = AnsiColors{
	Reset:  "\033[0m",
	Red:    "\033[31m",
	Green:  "\033[32m",
	Yellow: "\033[33m",
	Blue:   "\033[34m",
	Purple: "\033[35m",
	Cyan:   "\033[36m",
	White:  "\033[37m",
}

func (a AnsiColors) BeautifyCard(card ai.AnkiCard) string {
	return fmt.Sprintf(
		"%sFront:%s %s%s\n%sBack:%s %s%s\n\n",
		Colors.Blue, Colors.Reset,
		card.Front, Colors.Reset,
		Colors.Green, Colors.Reset,
		card.Back, Colors.Reset,
	)
}

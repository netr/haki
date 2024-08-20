package cmd

import "errors"

var (
	ErrWordFlagRequired = errors.New("word is required --word <word>")
)

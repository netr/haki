package cmd

import (
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v2"
)

type ErrFlagValueMissing struct {
	Flag string
}

func (e *ErrFlagValueMissing) Error() string {
	return fmt.Sprintf("flag '%s' is missing data", e.Flag)
}

type ActionCallbackFunc func() error

type Actioner interface {
	Run(...interface{}) error
	Flags() []string
	Name() string
}

type Action struct {
	flags  []string
	apiKey string
	name   string
}

func NewAction(apiKey, name string, flags []string) *Action {
	return &Action{
		flags:  flags,
		apiKey: apiKey,
		name:   name,
	}
}

func (a Action) Flags() []string {
	return a.flags
}

func (a Action) Name() string {
	return a.name
}

func actionFn(a Actioner) func(cCtx *cli.Context) error {
	return func(cCtx *cli.Context) error {
		var args []interface{}
		for _, f := range a.Flags() {
			fstr := cCtx.String(f)
			if fstr == "" {
				return &ErrFlagValueMissing{Flag: f}
			}
			args = append(args, cCtx.String(f))
		}

		if err := a.Run(args...); err != nil {
			slog.Error(
				"run",
				slog.String("action", a.Name()),
				slog.String("error", err.Error()),
			)
			return err
		}
		return nil
	}
}

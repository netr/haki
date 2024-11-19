package cmd

import (
	"log/slog"

	"github.com/urfave/cli/v2"
)

type ActionCallbackFunc func() error

type Actioner interface {
	Run(...interface{}) error
	Flags() []string
	Name() string
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
			slog.Error("run", slog.String("action", a.Name()), slog.String("error", err.Error()))
			return err
		}
		return nil
	}
}

func generateAnkiCardPrompt() string {
	prompt := `<Examples>
  <Documents>
    <Document>
      <Front>What is the capital of France?</Front>
      <Back>Paris</Back>
    </Document>
    <Document>
      <Front>What is a catalyst? (noun)</Front>
      <Back>A substance that increases the rate of a chemical reaction without itself undergoing any permanent chemical change.
      
	  Example: A “runaway feedback loop” describes a situation in which the output of a reaction becomes its own catalyst (auto-catalysis).
	  </Back>
    </Document>
    <Document>
      <Front>What is a sobriquet? (noun)</Front>
      <Back>A person's nickname or a descriptive name that is popularly used instead of the real name.
      
	  Example: The city has earned its sobriquet of 'the Big Apple'.
	  </Back
    </Document>
    <Document>
      <Front>How do you find the slope using the general form Ax + By = C?</Front>
      <Back>The slope is <anki-mathjax>-{A \\over B}</anki-mathjax></Back>
    </Document>
    <Document>
      <Front>What is a watershed moment? (noun)</Front>
      <Back>Zozobra is a feeling of anxiety or unease; the sensation that things are not as they should be or are on the brink of catastrophic failure.

	  Example: The constant updates of breaking news left her with a sense of zozobra, as she couldn't shake the feeling of impending doom.
	</Back>
    </Document>
    <Document>
      <Front>What is a watershed moment? (noun)</Front>
      <Back>A watershed moment is a critical turning point that signifies a major shift or change in direction. It's an event that causes significant and often transformative change, shaping the course of events thereafter.

    Examples:
    - The invention of the internet was a watershed moment in technology and communication.
    - The fall of the Berlin Wall marked a watershed moment in world history, symbolizing the end of the Cold War.

    Metaphor: Just as a watershed in geography is the line dividing waters flowing to different rivers or seas, a watershed moment in life represents a division between what came before and what follows.
	</Back>
    </Document>
    <Document>
      <Front>What are the four most common reasons an inequality sign must be reversed?</Front>
      <Back>
        The four most common reasons an inequality sign must be reversed are:
        - Multiplying or dividing both sides by a negative number: When you multiply or divide both sides of an inequality by a negative number, the inequality sign must be reversed.
        - Taking the reciprocal of both sides: If both sides of the inequality are positive and you take the reciprocal of each side, the inequality sign must be reversed.
        - Switching sides: If you swap the expressions on either side of the inequality, the inequality sign must be reversed to maintain the correct relationship.
        - Applying a decreasing function: When applying a function that is strictly decreasing (e.g., taking the logarithm of both sides in some cases), the inequality sign must be reversed.
      </Back>
    </Document>
  </Documents>
</Examples>
<Task>
Your task is to create insightful, meaningful and concise Anki cards with just a front and back. The back *MUST* be in HTML format. The goal is to create the most useful back cards as possible, to help the student learn deeply as they study. Please avoid wasteful and anemic questions.
- Prefer to use mathematical equations to explain the cards. Always wrap them in MathJax. Use """html <anki-mathjax>#MATH#</anki-mathjax>""". 
- When using code examples, *always* use Python.
- Wrap all code and psuedocode in <code></code>.
- Use a variety of methods when writing back card content, if it's helpful. I.e.: lists, concepts, examples, math equations, comparisons, usage, potential bias, etc.
- Always keep your answers concise without losing context and insight.
</Task>`

	return prompt
}

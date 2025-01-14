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

func generateAnkiCardPrompt() string {
	prompt := `<ankigen_examples>
  <Documents>
    <Document>
      <Front>What is the capital of France?</Front>
      <Back>Paris</Back>
    </Document>
    <Document>
      <Front>What is paltry? (adjective)</Front>
      <Back><div>Insignificant or meager; lacking in importance or worth.<br><br><b>Example:</b> The company's paltry profits were not enough to cover its expenses.<br><br><b>Synonyms:</b> trivial, negligible, meager, insignificant.</div></Back>
    </Document>
    <Document>
      <Front>What is insipid? (adjective)</Front>
      <Back><div>Lacking flavor, vigor, or interest; dull or boring.<br><br><b>Example:</b> The soup was rather insipid, lacking any real taste or seasoning.<br><br><b>Synonyms:</b> bland, tasteless, uninteresting, dull.</div></Back>
    </Document>
    <Document>
      <Front>How do you find the slope using the general form Ax + By = C?</Front>
      <Back>The slope is <anki-mathjax>-{A \\over B}</anki-mathjax></Back>
    </Document>
    <Document>
      <Front>What is Zozobra? (noun)</Front>
      <Back>Feeling of anxiety or unease; the sensation that things are not as they should be or are on the brink of catastrophic failure.<br><br><b>Example:</b> The constant updates of breaking news left her with a sense of zozobra, as she couldn't shake the feeling of impending doom.</Back>
    </Document>
    <Document>
      <Front>What is a watershed moment? (noun)</Front>
      <Back>A critical turning point that signifies a major shift or change in direction. It's an event that causes significant and often transformative change, shaping the course of events thereafter.<br><b>Examples:</b><br>- The invention of the internet was a watershed moment in technology and communication.<br>- The fall of the Berlin Wall marked a watershed moment in world history, symbolizing the end of the Cold War.<br><br><b>Metaphor:</b> Just as a watershed in geography is the line dividing waters flowing to different rivers or seas, a watershed moment in life represents a division between what came before and what follows.</Back>
    </Document>
    <Document>
      <Front>What are the four most common reasons an inequality sign must be reversed?</Front>
      <Back><div>The four most common reasons an inequality sign must be reversed are:<br>1. Multiplying or dividing both sides by a negative number: When you multiply or divide both sides of an inequality by a negative number, the inequality sign must be reversed.<br>2. Taking the reciprocal of both sides: If both sides of the inequality are positive and you take the reciprocal of each side, the inequality sign must be reversed.<br>3. Switching sides: If you swap the expressions on either side of the inequality, the inequality sign must be reversed to maintain the correct relationship.<br>4. Applying a decreasing function: When applying a function that is strictly decreasing (e.g., taking the logarithm of both sides in some cases), the inequality sign must be reversed.</Back>
    </Document>
    <Document>
      <Front>What are the six trigonometric functions?</Front>
      <Back><ul><li>Sine (sin): <anki-mathjax>\sin(\theta) = \frac{\text{opposite}}{\text{hypotenuse}}</anki-mathjax>=&nbsp;<anki-mathjax>y \over r</anki-mathjax></li><li>Cosine (cos): <anki-mathjax>\cos(\theta) = \frac{\text{adjacent}}{\text{hypotenuse}}</anki-mathjax>=&nbsp;<anki-mathjax>x \over r</anki-mathjax></li><li>Tangent (tan): <anki-mathjax>\tan(\theta) = \frac{\text{opposite}}{\text{adjacent}}</anki-mathjax>&nbsp;=&nbsp;<anki-mathjax> y \over x</anki-mathjax></li><li>Cotangent (cot):&nbsp;<anki-mathjax>\cot(\theta) = \frac{1}{\tan(\theta)}</anki-mathjax>&nbsp;=&nbsp;<anki-mathjax>x \over y</anki-mathjax></li><li>Secant (sec):&nbsp;<anki-mathjax>\sec(\theta) = \frac{1}{\cos(\theta)}</anki-mathjax>&nbsp;=&nbsp;<anki-mathjax>r \over x</anki-mathjax></li><li>Cosecant (csc): <anki-mathjax>\csc(\theta) = \frac{1}{\sin(\theta)}</anki-mathjax>&nbsp;=&nbsp;<anki-mathjax>r \over y</anki-mathjax></li></ul></Back>  
 	</Document>
    <Document>
      <Front>What is Normalized Discounted Cumulative Gain (NDCG) and why use it?</Front>
      <Back><div>A metric used to evaluate the performance of ranking algorithms, particularly in information retrieval, search engines, and recommendation systems.</div><h3><strong>Key Concepts</strong>:</h3><ol><li><div><strong>Purpose</strong>:</div><ul><li>Measures the quality of the ranking produced by an algorithm relative to an ideal ranking.</li><li>Accounts for both the relevance of results and their order in the list.</li></ul></li><li><div><strong>Discounting</strong>:</div><ul><li>Assigns higher importance to relevant items appearing earlier in the ranking.</li><li>Uses a logarithmic scale to reduce the impact of lower-ranked items.</li></ul></li><li><div><strong>Normalization</strong>:</div><ul><li>Ensures that scores are comparable across queries by dividing the raw DCG by the ideal DCG (i.e., the best possible ranking).</li><li>Produces a value between 0 and 1, where 1 represents a perfect ranking.</li></ul></li></ol></Back>
    </Document>
  </Documents>
</ankigen_examples>
<ankigen_info>
    AnkiGen is an advanced AI anki card generatiion assistant created by Netr.
    AnkiGen is designed to emulate the world's most proficient learners.
    AnkiGen is always up-to-date with the latest note taking/flash card skills and best practices.
    AnkiGen responds with back ankiCards that use HTML format.
	AnkiGen wraps all code and psuedocode in <code></code>.
	AnkiGen only writes code in Python.
	AnkiGen prefers using mathematical equations to explain the ankiCards. Always wrap them in MathJax. Use """html <anki-mathjax>#MATH#</anki-mathjax>""". 
    AnkiGen aims to deliver clear, concise, and effective flash ankiCards while maintaining an engaging and entertaining manner.
  
    AnkiGen's knowledge spans various disciplines but emphasizes mathematics and computer science when applicable.
  </ankigen_info>`

	return prompt
}

// nolint: unused
func generateAnkiCardPrompt_V1() string {
	prompt := `<Examples>
  <Documents>
    <Document>
      <Front>What is the capital of France?</Front>
      <Back>Paris</Back>
    </Document>
    <Document>
      <Front>What is a catalyst? (noun)</Front>
      <Back>A substance that increases the rate of a chemical reaction without itself undergoing any permanent chemical change.<br><br><b>Example:</b> A “runaway feedback loop” describes a situation in which the output of a reaction becomes its own catalyst (auto-catalysis).</Back>
    </Document>
    <Document>
      <Front>What is a sobriquet? (noun)</Front>
      <Back>A person's nickname or a descriptive name that is popularly used instead of the real name.<br><br><b>Example:</b> The city has earned its sobriquet of 'the Big Apple'.</Back
    </Document>
    <Document>
      <Front>How do you find the slope using the general form Ax + By = C?</Front>
      <Back>The slope is <anki-mathjax>-{A \\over B}</anki-mathjax></Back>
    </Document>
    <Document>
      <Front>What is a watershed moment? (noun)</Front>
      <Back>Zozobra is a feeling of anxiety or unease; the sensation that things are not as they should be or are on the brink of catastrophic failure.<br><br><b>Example:</b> The constant updates of breaking news left her with a sense of zozobra, as she couldn't shake the feeling of impending doom.</Back>
    </Document>
    <Document>
      <Front>What is a watershed moment? (noun)</Front>
      <Back>A watershed moment is a critical turning point that signifies a major shift or change in direction. It's an event that causes significant and often transformative change, shaping the course of events thereafter.<br><b>Examples:</b><br>- The invention of the internet was a watershed moment in technology and communication.<br>- The fall of the Berlin Wall marked a watershed moment in world history, symbolizing the end of the Cold War.<br><br><b>Metaphor:</b> Just as a watershed in geography is the line dividing waters flowing to different rivers or seas, a watershed moment in life represents a division between what came before and what follows.</Back>
    </Document>
    <Document>
      <Front>What are the four most common reasons an inequality sign must be reversed?</Front>
      <Back><div>The four most common reasons an inequality sign must be reversed are:<br>1. Multiplying or dividing both sides by a negative number: When you multiply or divide both sides of an inequality by a negative number, the inequality sign must be reversed.<br>2. Taking the reciprocal of both sides: If both sides of the inequality are positive and you take the reciprocal of each side, the inequality sign must be reversed.<br>3. Switching sides: If you swap the expressions on either side of the inequality, the inequality sign must be reversed to maintain the correct relationship.<br>4. Applying a decreasing function: When applying a function that is strictly decreasing (e.g., taking the logarithm of both sides in some cases), the inequality sign must be reversed.</Back>
    </Document>
    <Document>
      <Front>What are the six trigonometric functions?</Front>
      <Back><ul><li>Sine (sin): <anki-mathjax>\sin(\theta) = \frac{\text{opposite}}{\text{hypotenuse}}</anki-mathjax>=&nbsp;<anki-mathjax>y \over r</anki-mathjax></li><li>Cosine (cos): <anki-mathjax>\cos(\theta) = \frac{\text{adjacent}}{\text{hypotenuse}}</anki-mathjax>=&nbsp;<anki-mathjax>x \over r</anki-mathjax></li><li>Tangent (tan): <anki-mathjax>\tan(\theta) = \frac{\text{opposite}}{\text{adjacent}}</anki-mathjax>&nbsp;=&nbsp;<anki-mathjax> y \over x</anki-mathjax></li><li>Cotangent (cot):&nbsp;<anki-mathjax>\cot(\theta) = \frac{1}{\tan(\theta)}</anki-mathjax>&nbsp;=&nbsp;<anki-mathjax>x \over y</anki-mathjax></li><li>Secant (sec):&nbsp;<anki-mathjax>\sec(\theta) = \frac{1}{\cos(\theta)}</anki-mathjax>&nbsp;=&nbsp;<anki-mathjax>r \over x</anki-mathjax></li><li>Cosecant (csc): <anki-mathjax>\csc(\theta) = \frac{1}{\sin(\theta)}</anki-mathjax>&nbsp;=&nbsp;<anki-mathjax>r \over y</anki-mathjax></li></ul></Back>  
 	</Document>
    <Document>
      <Front>What is Normalized Discounted Cumulative Gain (NDCG) and why use it?</Front>
      <Back><div>A metric used to evaluate the performance of ranking algorithms, particularly in information retrieval, search engines, and recommendation systems.</div><h3><strong>Key Concepts</strong>:</h3><ol><li><div><strong>Purpose</strong>:</div><ul><li>Measures the quality of the ranking produced by an algorithm relative to an ideal ranking.</li><li>Accounts for both the relevance of results and their order in the list.</li></ul></li><li><div><strong>Discounting</strong>:</div><ul><li>Assigns higher importance to relevant items appearing earlier in the ranking.</li><li>Uses a logarithmic scale to reduce the impact of lower-ranked items.</li></ul></li><li><div><strong>Normalization</strong>:</div><ul><li>Ensures that scores are comparable across queries by dividing the raw DCG by the ideal DCG (i.e., the best possible ranking).</li><li>Produces a value between 0 and 1, where 1 represents a perfect ranking.</li></ul></li></ol></Back>
    </Document>
  </Documents>
</Examples>
<Task>
Your task is to create insightful, meaningful and concise Anki ankiCards with just a front and back. The goal is to create the most useful back ankiCards as possible, to help the student learn deeply as they study. Please avoid wasteful and anemic questions.
- Prefer to use mathematical equations to explain the ankiCards. Always wrap them in MathJax. Use """html <anki-mathjax>#MATH#</anki-mathjax>""". 
- When using code examples, *always* use Python.
- Wrap all code and psuedocode in <code></code>.
- Use a variety of methods when writing back card content, if it's helpful. I.e.: lists, concepts, examples, math equations, comparisons, usage, potential bias, etc.
- Always keep your answers concise without losing context and insight.
- The back card *MUST only* be in HTML format!
</Task>`

	return prompt
}

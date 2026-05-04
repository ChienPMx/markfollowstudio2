package whispercpp

type WhispercppProcessor struct {
	WorkDir string // Directory for generating intermediate files
	Model   string
}

func NewWhispercppProcessor(model string) *WhispercppProcessor {
	return &WhispercppProcessor{
		Model: model,
	}
}

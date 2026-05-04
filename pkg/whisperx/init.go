package whisperx

type WhisperXProcessor struct {
	WorkDir string // Directory for generating intermediate files
	Model   string
}

func NewWhisperXProcessor(model string) *WhisperXProcessor {
	return &WhisperXProcessor{
		Model: model,
	}
}

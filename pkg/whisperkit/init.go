package whisperkit

type WhisperKitProcessor struct {
	WorkDir string // Directory for generating intermediate files
	Model   string
}

func NewWhisperKitProcessor(model string) *WhisperKitProcessor {
	return &WhisperKitProcessor{
		Model: model,
	}
}

package fasterwhisper

type FastwhisperProcessor struct {
	WorkDir string // Directory for generating intermediate files
	Model   string
}

func NewFastwhisperProcessor(model string) *FastwhisperProcessor {
	return &FastwhisperProcessor{
		Model: model,
	}
}

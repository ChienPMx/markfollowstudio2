package util

import (
	"go.uber.org/zap"
	"markflow-studio/internal/storage"
	"markflow-studio/log"
	"os/exec"
	"path/filepath"
	"strings"
)

// Process audio to mono, 16k sample rate
func ProcessAudio(filePath string) (string, error) {
	dest := strings.ReplaceAll(filePath, filepath.Ext(filePath), "_mono_16K.mp3")
	cmdArgs := []string{"-i", filePath, "-ac", "1", "-ar", "16000", "-b:a", "192k", dest}
	cmd := exec.Command(storage.FfmpegPath, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.GetLogger().Error("Audio processing failed", zap.Error(err), zap.String("audio file", filePath), zap.String("output", string(output)))
		return "", err
	}
	return dest, nil
}

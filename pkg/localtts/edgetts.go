package localtts

import (
	"context"
	"fmt"
	"io/ioutil"
	"krillin-ai/internal/storage"
	"krillin-ai/log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

type EdgeTtsClient struct {
}

func NewEdgeTtsClient() *EdgeTtsClient {
	return &EdgeTtsClient{}
}

func (c *EdgeTtsClient) Text2Speech(text, voice, outputFile string) error {
	// Clean extra spaces in voice names
	voice = strings.TrimSpace(voice)
	if voice == "" {
		voice = config.Conf.Tts.EdgeTts.Voice
	}
	if voice == "" {
		voice = "en-US-AndrewNeural"
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.GetLogger().Error("Failed to create output directory", zap.String("dir", outputDir), zap.Error(err))
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get absolute path
	absOutputFile, err := filepath.Abs(outputFile)
	if err != nil {
		log.GetLogger().Error("Failed to get absolute path of output file", zap.Error(err))
		return fmt.Errorf("failed to get absolute path of output file: %w", err)
	}
	absOutputDir := filepath.Dir(absOutputFile)

	// Create temporary file to store text content to avoid command line argument escaping issues
	tempFile, err := os.CreateTemp("", "edge-tts-*.txt")
	if err != nil {
		log.GetLogger().Error("Failed to create temporary file", zap.Error(err))
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempFileName := tempFile.Name()

	// Ensure temporary file is cleaned up at the end
	defer func() {
		tempFile.Close()
		if err := os.Remove(tempFileName); err != nil {
			log.GetLogger().Warn("Failed to clean up temporary file", zap.String("file", tempFileName), zap.Error(err))
		}
	}()

	// Write text to temporary file
	if _, err := tempFile.WriteString(text); err != nil {
		log.GetLogger().Error("Failed to write to temporary file", zap.Error(err))
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}
	tempFile.Close() // Ensure file is written

	// Retry mechanism
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.GetLogger().Info("edge-tts transcription attempt",
			zap.Int("attempt", attempt),
			zap.Int("maxRetries", maxRetries),
			zap.String("text_length", fmt.Sprintf("%d", len(text))))

		err := c.attemptTTS(tempFileName, voice, absOutputFile, attempt)
		if err == nil {
			// Successfully generated
			log.GetLogger().Info("edge-tts transcription complete", zap.String("output file", absOutputFile))
			if _, err := os.Stat(absOutputFile); os.IsNotExist(err) {
				log.GetLogger().Error("edge-tts output file does not exist", zap.String("output file", absOutputFile))
				return fmt.Errorf("edge-tts output file does not exist: %s", absOutputFile)
			}
			return nil
		}

		log.GetLogger().Warn("edge-tts transcription failed, preparing to retry",
			zap.Int("attempt", attempt),
			zap.Error(err))

		// If not the last attempt, wait for a while before retrying
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * 2 * time.Second
			log.GetLogger().Info("Waiting for retry", zap.Duration("waitTime", waitTime))
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("edge-tts transcription failed after %d retries", maxRetries)
}

func (c *EdgeTtsClient) attemptTTS(tempFileName, voice, absOutputFile string, attempt int) error {
	// Use new edge-tts command parameters (file input mode)
	cmdArgs := []string{
		"--text-file", tempFileName,
		"--voice", voice,
		"--output", absOutputFile,
		"--format", "wav",
		"--sample_rate", "44100",
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // 60s timeout
	defer cancel()

	cmd := exec.CommandContext(ctx, storage.EdgeTtsPath, cmdArgs...)
	log.GetLogger().Info("edge-tts transcription started",
		zap.String("cmd", cmd.String()),
		zap.String("temp_file", tempFileName),
		zap.String("output_file", absOutputFile),
		zap.Int("attempt", attempt))

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.GetLogger().Error("edge-tts cmd timeout", zap.String("output", string(output)), zap.Error(err))
			return fmt.Errorf("edge-tts execution timeout")
		}
		log.GetLogger().Error("edge-tts cmd execution failed", zap.String("output", string(output)), zap.Error(err))
		return err
	}

	return nil
}
